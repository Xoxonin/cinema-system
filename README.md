# Cinema System - Mikroserwisy, Docker i Docker Compose

Projekt prezentuje system rezerwacji kinowych oparty o mikroserwisy:

- `user-service`
- `catalog-service`
- `showtime-service`
- `booking-service`
- `frontend`

Całość uruchamiana jest przez Docker Compose.

## Architektura

System składa się z pięciu głównych komponentów: frontendu oraz czterech mikroserwisów backendowych. Każdy z mikroserwisów odpowiada za osobną domenę biznesową i posiada własną, niezależną bazę danych PostgreSQL.

- **Frontend**: Aplikacja React + Vite serwowana przez zabezpieczony serwer Nginx.
- **Go Microservices**: Cztery wydajne serwisy backendowe napisane w języku Go (`user-service`, `catalog-service`, `showtime-service`, `booking-service`).
- **PostgreSQL StatefulSets**: Dedykowane bazy danych dla każdego serwisu, gwarantujące pełną izolację danych.
- **Db Migrations**: Kontenery automatycznie wykonujące migracje schematów baz danych przy starcie.

### Schemat Działania i Przepływu Ruchu w Klastrze Kubernetes

Poniższy schemat Mermaid przedstawia pełną architekturę sieciową systemu w Kubernetes, obrazując trasowanie żądań zewnętrznych przez Ingress oraz alternatywną ścieżkę NodePort, a także komunikację między przestrzeniami nazw (`frontend-ns` i `backend-ns`):

```mermaid
graph TD
    %% Entry Points
    subgraph Klienci i Brzeg Klastra [Dostęp Zewnętrzny]
        Client[Przeglądarka Klienta]
        Ingress[cinema-ingress Ingress - Port 80 / 443]
        NodePort[Frontend NodePort Service - Port 30080]
    end

    %% frontend-ns
    subgraph frontend-ns [Przestrzeń nazw: frontend-ns]
        FrontendPod[frontend-pod: React + Nginx Proxy]
        
        subgraph ExternalNames [Usługi Mapujące CNAME]
            ExtUser[user-service ExternalName]
            ExtCatalog[catalog-service ExternalName]
            ExtShowtime[showtime-service ExternalName]
            ExtBooking[booking-service ExternalName]
        end
    end

    %% backend-ns
    subgraph backend-ns [Przestrzeń nazw: backend-ns]
        subgraph GoServices [Mikroserwisy Go]
            UserService[user-service pod:8081]
            CatalogService[catalog-service pod:8082]
            ShowtimeService[showtime-service pod:8083]
            BookingService[booking-service pod:8084]
        end

        subgraph PostgreSQL [Trwała Warstwa Danych]
            DbUsers[(db-users StatefulSet:5432)]
            DbCatalog[(db-catalog StatefulSet:5432)]
            DbShowtime[(db-showtime StatefulSet:5432)]
            DbBooking[(db-booking StatefulSet:5432)]
        end
    end

    %% Connections - Ingress Path
    Client -->|Żądanie HTTP| Ingress
    
    Ingress -->|Ścieżka / | FrontendPod
    Ingress -->|Ścieżka /api/users | ExtUser
    Ingress -->|Ścieżka /api/movies | ExtCatalog
    Ingress -->|Ścieżka /api/showtimes | ExtShowtime
    Ingress -->|Ścieżka /api/rooms | ExtShowtime
    Ingress -->|Ścieżka /api/bookings | ExtBooking

    %% Connections - NodePort Path
    Client -.->|Alternatywny dostęp NodePort| NodePort
    NodePort --> FrontendPod
    FrontendPod -.->|Proxy dla /api/* bez Ingress| GoServices

    %% ExternalNames Routing
    ExtUser --> UserService
    ExtCatalog --> CatalogService
    ExtShowtime --> ShowtimeService
    ExtBooking --> BookingService

    %% Services to Databases
    UserService --> DbUsers
    CatalogService --> DbCatalog
    ShowtimeService --> DbShowtime
    BookingService --> DbBooking

    %% Styling
    style Client fill:#dec0f1,stroke:#333,stroke-width:2px,color:#000
    style Ingress fill:#ff69b4,stroke:#333,stroke-width:2px,color:#000
    style NodePort fill:#f4a261,stroke:#333,stroke-width:2px,color:#000
    style FrontendPod fill:#6495ed,stroke:#333,stroke-width:2px,color:#000
    
    style ExtUser fill:#e9c46a,stroke:#333,stroke-width:1px,color:#000
    style ExtCatalog fill:#e9c46a,stroke:#333,stroke-width:1px,color:#000
    style ExtShowtime fill:#e9c46a,stroke:#333,stroke-width:1px,color:#000
    style ExtBooking fill:#e9c46a,stroke:#333,stroke-width:1px,color:#000

    style UserService fill:#90ee90,stroke:#333,stroke-width:2px,color:#000
    style CatalogService fill:#90ee90,stroke:#333,stroke-width:2px,color:#000
    style ShowtimeService fill:#90ee90,stroke:#333,stroke-width:2px,color:#000
    style BookingService fill:#90ee90,stroke:#333,stroke-width:2px,color:#000

    style DbUsers fill:#ff6b6b,stroke:#333,stroke-width:2px,color:#000
    style DbCatalog fill:#ff6b6b,stroke:#333,stroke-width:2px,color:#000
    style DbShowtime fill:#ff6b6b,stroke:#333,stroke-width:2px,color:#000
    style DbBooking fill:#ff6b6b,stroke:#333,stroke-width:2px,color:#000
```


## Uruchomienie lokalne (dev)

```powershell
docker compose up -d --build
docker compose ps
```

Frontend jest dostępny pod adresem:

- `http://localhost:8080`

## Wykonanie wymagań (punkty 2-6)

### Punkt 2: Dockerfile i dobre praktyki

Zrealizowano Dockerfile dla wszystkich zaplanowanych mikroserwisów i frontendu, z uwzględnieniem:

- multi-stage build
- oddzielenia etapu build/runtime
- uruchamiania jako non-root (backendy)
- plików `.dockerignore` dla kontekstów builda

Zrzut ekranu Dockerfile dla `catalog-service`:

![Catalog service Dockerfile](docs/artifacts/dockerfile%20catalog-service.png)

### Punkt 3: Multiarch + DockerHub + SBOM

Build i push obrazów przygotowany przez `Makefile`:

- `docker buildx build --platform linux/amd64,linux/arm64`
- `--sbom=true`
- `--push` (dla publikacji do DockerHub)

Użyte targety:

```powershell
make build-all REPO=<dockerhub_namespace> VERSION=<tag>
make push-all REPO=<dockerhub_namespace> VERSION=<tag>
```

Zrzuty ekranu potwierdzajace wykonanie punktu 3:

- Fragment `Makefile` z targetami buildx (`linux/amd64,linux/arm64`, `--sbom=true`, `--push`):

![Makefile buildx](docs/artifacts/screen_makefile.png)

- Widok repozytoriow obrazow na DockerHub:

![DockerHub repositories](docs/artifacts/dockerhub_wszystkie.png)

- Szczegoly obrazu `frontend` na DockerHub (manifest list z dwiema architekturami):

![Frontend multiarch details](docs/artifacts/dockerhub_frontend_szczegoly.png)

Na powyzszym zrzucie widac, ze obraz `frontend` jest wieloarchitekturowy i zawiera `linux/amd64` oraz `linux/arm64`.

### Punkt 4: Analiza podatności (Trivy)

Wykonano skany obrazów.

Wynik: wszystkie wykonane skany wykazały `0` podatności (`0` HIGH i `0` CRITICAL).

Zrzuty potwierdzające:

- ![Trivy user-service](docs/artifacts/Skan%20user-service.png)
- ![Trivy catalog-service](docs/artifacts/Skan%20catalog-service.png)
- ![Trivy showtime-service](docs/artifacts/Skan%20showtime-service.png)
- ![Trivy booking-service](docs/artifacts/Skan%20booking-service.png)
- ![Trivy frontend](docs/artifacts/Skan%20frontend.png)

### Punkt 5: Deweloperski docker-compose

Plik `docker-compose.yaml` zawiera praktyki wspierające utrzymanie i uruchamianie projektu:

Celowo pominieto pole `version`, poniewaz w Docker Compose V2 obecny silnik je ignoruje (pole jest przestarzale/obsolete).

- jawnie zdefiniowaną sieć (`cinema-net`)
- named volumes dla trwałości danych
- healthcheck dla baz danych
- `depends_on` z warunkami startu
- limity zasobów (`deploy.resources.limits`)
- polityki restartu (`restart: unless-stopped`, migracje: `restart: "no"`)

Zrzut ekranu pokazujacy poczatek konfiguracji Compose:

![Docker Compose structure](docs/artifacts/docker%20compose.png)

### Punkt 6: Graficzna reprezentacja aplikacji

Reprezentacja została wygenerowana narzędziem `compose-viz` i zapisana jako:

- `docs/artifacts/compose-viz.svg`

![Compose Viz](docs/artifacts/compose-viz.svg)

---

# Zadanie projektowe 2

## Zaawansowana, deklaratywna architektura Kubernetes dla projektu Cinema System

W tej sekcji opisano zaawansowaną architekturę orkiestracji kontenerów w wielowęzłowym środowisku Kubernetes, wdrożoną w projekcie **cinema-system** w celach zapewnienia wysokiej niezawodności, skalowalności i pełnego bezpieczeństwa chmurowego.

Szczegółowy opis techniczny zaawansowanych mechanizmów bezpieczeństwa, limitów oraz reguł orkiestracji znajduje się również w dedykowanym pliku: **bezpieczenstwo_i_mechanizmy.md**

---

### 1. Podział na przestrzenie nazw (Namespaces)

W celu zachowania izolacji architektonicznej i ograniczenia uprawnień poszczególnych komponentów (zasada najmniejszych uprawnień), system został rozdzielony na dwie dedykowane, niestandardowe przestrzenie nazw:

*   **`frontend-ns` (Przestrzeń użytkownika i routerów):**
    Tutaj wdrażany jest publicznie dostępny frontend aplikacji oraz komponenty brzegowe sieci (Ingress Controller, Ingress, HPA, usługi typu `ExternalName` mapujące ruch). Wydzielenie tej strefy zapobiega bezpośredniemu dostępowi z zewnątrz klastra do newralgicznych komponentów bazodanowych w przypadku naruszenia bezpieczeństwa warstwy brzegowej.
*   **`backend-ns` (Przestrzeń logiki biznesowej i danych):**
    Przeznaczona dla 4 podstawowych mikrousług Go (`user-service`, `catalog-service`, `showtime-service`, `booking-service`) oraz 4 dedykowanych baz danych PostgreSQL. Ruch sieciowy w tej przestrzeni jest całkowicie odcięty od świata zewnętrznego i podlega ścisłej kontroli za pomocą polis sieciowych.

---

### 2. Wybór kontrolerów obciążeń (Workload Controllers)

Dobór odpowiednich kontrolerów w Kubernetes ma fundamentalne znaczenie dla odporności na awarie i zarządzania stanem:

*   **Deployment (Mikrousługi bezstanowe: `user`, `catalog`, `showtime`, `booking` oraz `frontend`):**
    Wszystkie aplikacje Go oraz serwer proxy Nginx są z natury bezstanowe. Wybrano dla nich kontroler `Deployment`, który zapewnia:
    *   **Skalowanie poziome:** Uruchomiono po **2 repliki** dla każdego backendu, co gwarantuje wysoką dostępność. Jeśli jeden pod ulegnie awarii, drugi natychmiast przejmuje ruch.
    *   **RollingUpdate (Bezprzestojowość):** Aktualizacje kodu odbywają się z zerowym czasem przestoju. Wdrożone parametry `maxSurge: 25%` oraz `maxUnavailable: 25%` (wyrażone procentowo) oznaczają, że w trakcie wdrażania nowej wersji systemu co najmniej 75% starych replik ciągle działa, a nowe pody są stopniowo włączane po przejściu testów gotowości (`readinessProbe`).
*   **StatefulSet (Bazy danych PostgreSQL: `db-users`, `db-catalog`, `db-showtime`, `db-booking`):**
    Bazy danych przechowują stan systemu. Stosowanie zwykłego `Deployment` groziłoby uszkodzeniem danych z powodu braku gwarancji kolejności uruchamiania i tożsamości sieciowej podów. Zastosowano kontroler `StatefulSet` z **1 repliką** dla każdej bazy danych, co zapewnia:
    *   **Unikalna tożsamość:** Każdy pod bazy danych otrzymuje stałą nazwę (np. `db-users-0`) i stały dysk, który nie zmienia się przy restartach.
    *   **volumeClaimTemplates:** Umożliwia stabilne, automatyczne powiązanie każdego podu ze swoim dedykowanym wolumenem PersistentVolume.
*   **DaemonSet (Uzasadnienie braku wyboru):**
    Kontroler `DaemonSet` uruchamia dokładnie jedną replikę poda na każdym węźle klastra. Jest on idealny dla agentów systemowych, takich jak Cilium CNI, zbieranie logów (Fluentd) czy monitoring (Prometheus Node Exporter). Ponieważ poszczególne mikroserwisy biznesowe powinny być dynamicznie harmonogramowane i skalowane przez Kubernetes w zależności od rzeczywistego obciążenia (a nie uruchamiane "na sztywno" na każdym fizycznym węźle bez względu na zasoby), **celowo nie wybrano kontrolera DaemonSet** dla żadnego z komponentów aplikacji.

---

### 3. Konfiguracja usług sieciowych (Services)

Wdrożono i zademonstrowano dwa podstawowe rodzaje abstrakcji usług sieciowych:

*   **ClusterIP (Wewnętrzna izolacja sieci):**
    Zastosowany dla wszystkich 4 mikrousług Go oraz ich baz danych wewnątrz przestrzeni `backend-ns`. Pody te nie posiadają zewnętrznych adresów IP i komunikują się wewnątrz sieci klastra za pomocą nazw domenowych CoreDNS o strukturze: `<nazwa-uslugi>.<przestrzen-nazw>.svc.cluster.local` (np. `db-users.backend-ns.svc.cluster.local`). Całkowicie eliminuje to twardo kodowane adresy IP z kodu aplikacji.
*   **NodePort (Ekspozycja brzegu klastra):**
    Frontend serwujący aplikację React został wystawiony za pomocą usługi typu `NodePort` (przypisany stały port `30080` na węzłach fizycznych klastra). Umożliwia to bezpośredni dostęp do serwera Nginx na adresie IP dowolnego węzła klastra Minikube, co jest idealne jako alternatywne wejście lub port diagnostyczny brzegu sieci.

---

### 4. Dostęp z zewnątrz klastra (Ingress)

Kierowanie ruchem publicznym HTTP z zewnątrz klastra realizowane jest przez deklaratywny zasób **Ingress** (z kontrolerem `ingress-nginx`) działający w przestrzeni `frontend-ns`:

*   **Path-based Routing (Trasowanie ścieżek HTTP):**
    Ingress analizuje adres URL zapytania przychodzącego z zewnątrz klastra i przekazuje ruch zgodnie z regułami:
    *   Ścieżka `/api/users` -> Usługa `user-service`
    *   Ścieżka `/api/movies` -> Usługa `catalog-service`
    *   Ścieżka `/api/showtimes` i `/api/rooms` -> Usługa `showtime-service`
    *   Ścieżka `/api/bookings` -> Usługa `booking-service`
    *   Ścieżka `/` -> Usługa `frontend` (serwująca pliki statyczne HTML/JS)
*   **Mapowanie Cross-Namespace (`ExternalName`):**
    Ponieważ Ingress może kierować ruch tylko do usług w tej samej przestrzeni nazw (`frontend-ns`), wdrożono wzorzec usług `ExternalName`. W `frontend-ns` utworzono usługi o identycznych nazwach jak te w backendzie, które przy zapytaniu zwracają rekord CNAME wskazujący na pełny adres mikrousługi w przestrzeni `backend-ns`. Dzięki temu routing na poziomie Ingress działa bezbłędnie i bez konieczności duplikowania podów.
*   **Ingress vs LoadBalancer / Gateway API (Uzasadnienie wyboru):**
    Wybrano mechanizm `Ingress` z kontrolerem `ingress-nginx` zamiast oddzielnych usług typu `LoadBalancer` dla każdego mikroserwisu. Usługa `LoadBalancer` tworzy osobny publiczny adres IP dla każdego serwisu (co w środowiskach chmurowych/wielowęzłowych generuje wysokie koszty i komplikuje zarządzanie). `Ingress` pozwala na skonsolidowanie całego ruchu pod **jednym publicznym adresem IP** i inteligentne trasowanie na podstawie ścieżek URL (path-based routing), co jest optymalne kosztowo i operacyjnie. Nowy standard `Gateway API` pominięto z uwagi na dodatkowy narzut konfiguracyjny (standard Ingress jest w pełni wystarczający i sprawdzony dla tej architektury).

---

### 5. Trwałość danych (PV, PVC, StorageClass)

Aby bazy danych PostgreSQL nie traciły danych podczas restartu podów, zaprojektowano wydajną warstwę składowania danych opartą na fizycznych dyskach węzłów:

*   **StorageClass (`local-storage`):**
    Klasa przechowywania danych wykorzystuje opcję `volumeBindingMode: WaitForFirstConsumer` (opóźnione wiązanie wolumenu). Sprawia to, że Kubernetes wstrzymuje się z przydziałem fizycznego wolumenu i powiązaniem z PVC do momentu, gdy planista (Scheduler) zdecyduje, na którym węźle fizycznym uruchomić dany pod bazy danych (uwzględniając reguły Affinity).
*   **Lokalne Wolumeny Persistent Volume (Local PV):**
    Dla każdej bazy danych zdefiniowano obiekt `PersistentVolume` z typem `local`, kierujący na fizyczny katalog na węźle `minikube` (np. `/mnt/data/db-users`). Zapewnia to maksymalną wydajność I/O (brak narzutu sieciowych dysków rozproszonych).
*   **PersistentVolumeClaim (PVC):**
    Generowane automatycznie z sekcji `volumeClaimTemplates` w specyfikacji `StatefulSet`. Gwarantuje to stałe, jednoznaczne powiązanie podu bazy danych ze swoim fizycznym dyskiem na konkretnym węźle.
*   **Blokada Wolumenów (`claimRef`):**
    Wszystkie cztery manifesty `PersistentVolume` zawierają sekcję `claimRef` wskazującą na konkretne roszczenie (PVC) w określonym namespace. To kluczowe zabezpieczenie chroni przed domyślnym, zachłannym przypisywaniem wolumenów lokalnych o tym samym rozmiarze do przypadkowych baz danych podczas restartu (np. sytuacja, w której baza użytkowników omyłkowo mountuje wolumen z danymi seansów kinowych).

---

### 6. Bezpieczne zarządzanie konfiguracją (ConfigMaps & Secrets)

Zgodnie z dobrymi praktykami 12-Factor App, oddzielono konfigurację aplikacji od kodu binarnego:

*   **ConfigMaps (Konfiguracja jawna):**
    Użyte do przechowywania plików migracji bazy danych `.sql` (np. `user-db-migrations`). Pliki te są deklaratywnie zapisane jako dane w obiekcie ConfigMap i montowane w kontenerach startowych (`initContainers`) jako wolumen w trybie tylko do odczytu, co umożliwia bezproblemowe przeprowadzenie migracji przy uruchomieniu aplikacji.
*   **Secrets (Dane wrażliwe):**
    Wszystkie hasła do baz danych, tokeny JWT (`JWT_SECRET`) oraz pełne parametry połączeń (`DATABASE_URL`) zostały odseparowane i wstrzyknięte poprzez obiekty typu `Secret`. Aplikacje Go pobierają je jako zmiennes środowiskowe zreferowane bezpośrednio w plikach wdrożeń (Deployment/StatefulSet), dzięki czemu hasła nigdy nie pojawiają się w repozytorium kodu.

---

### 7. Sondy kondycji i gotowości (Liveness & Readiness Probes)

Dla wszystkich pięciu mikroserwisów wdrożono mechanizmy samoleczenia (Self-healing) na poziomie kontenerów w celu automatycznej detekcji stanów zawieszenia oraz zapewnienia stabilności działania:

*   **Sonda Kondycji (Liveness Probe):**
    *   **Zastosowanie:** Zdefiniowana dla wszystkich aplikacji Go (porty `8081-8084`, ścieżka `/health`) oraz serwera Nginx frontendu.
    *   **Parametry:** `initialDelaySeconds: 10`, `periodSeconds: 10`.
    *   **Uzasadnienie:** Sonda cyklicznie odpytuje kontener. Jeśli aplikacja ulegnie zakleszczeniu (deadlock) lub wewnętrznej awarii, która uniemożliwi obsługę zapytań HTTP, Kubernetes automatycznie zrestartuje kontener, przywracając działanie usługi bez interwencji administratora.
*   **Sonda Gotowości (Readiness Probe):**
    *   **Zastosowanie:** Zdefiniowana dla aplikacji oraz frontendu (porty `8081-8084`, ścieżka `/health`).
    *   **Parametry:** `initialDelaySeconds: 5`, `periodSeconds: 5`.
    *   **Uzasadnienie:** Gwarantuje, że nowy pod nie otrzyma ruchu sieciowego od użytkowników przed pełnym zakończeniem inicjalizacji (np. zanim nawiąże stabilne połączenie z bazą danych i wczyta konfigurację). Jest to kluczowy element bezprzestojowego wdrażania (`RollingUpdate`) – stary pod jest wyłączany dopiero wtedy, gdy nowo utworzony pod zgłosi pełną gotowość.

---

### 8. Polityki sieciowe (Network Policies & Cilium CNI)

Zamiast domyślnej płaskiej sieci Kubernetes, wdrożono rygorystyczny, deklaratywny model bezpieczeństwa sieciowego z zasadą "Deny-by-Default" przy użyciu **CNI Cilium (eBPF)**:

*   **Domyślna odmowa (`default-deny-backend` / `default-deny-frontend`):**
    Całkowicie blokuje jakikolwiek nieautoryzowany ruch wejściowy i wyjściowy w klastrze dla obu przestrzeni nazw.
*   **Granularne zezwolenia:**
    *   **`allow-frontend`**: Zezwala na ruch przychodzący do frontendu na port 8080 (z Ingressa) oraz ruch wychodzący do DNS oraz mikrousług backendu na porty API (`8081-8084`).
    *   **`allow-backends`**: Zezwala na ruch wejściowy wyłącznie z przestrzeni `frontend-ns` oraz od **Ingress Controllera** (`ingress-nginx` / `kube-system`), zabezpieczając API i odcinając niepowołane podmioty klastra. Zezwala również na wyjście do DNS i baz danych.
    *   **`allow-databases`**: Blokuje jakikolwiek ruch wyjściowy (pełna ochrona przed wyciekiem danych) i przyjmuje ruch na porcie `5432` **wyłącznie** od odpowiadającego mu podu mikrousługi (np. `db-catalog` przyjmuje połączenia tylko z `catalog-service`).

---

### 9. Limity i przydziały zasobów (Resource Quotas & Limit Ranges)

Wdrożono zaawansowaną kontrolę zasobów fizycznych węzłów, chroniąc klaster przed przeciążeniem i atakami typu Denial of Service:

*   **LimitRange (`backend-limit-range`):**
    Definiuje domyślne żądania (`requests`) oraz limity (`limits`) dla CPU i pamięci RAM dla każdego kontenera, który sam ich nie określił, standaryzując środowisko uruchomieniowe.
*   **ResourceQuota (`backend-quota`):**
    Maksymalny sumaryczny "sufit" zasobów dla całej przestrzeni nazw. Zapobiega to przejęciu całości pamięci węzła przez jedną przestrzeń nazw (np. w przypadku niekontrolowanego autoskalowania HPA lub wycieku pamięci).
*   **Limity dla `db-migration`:**
    Aby uniknąć odrzucenia podów w środowiskach o rygorystycznych limitach ResourceQuota (np. gdy lokalne Minikube ma nieaktywny lub powolny mechanizm mutacji `LimitRange`), do kontenerów startowych `db-migration` dodano **jawne, niskie definicje zasobów** (50m CPU, 32Mi RAM), gwarantując bezbłędny start aplikacji w każdych warunkach.

---

### 10. Reguły planowania podów (Pod Affinity & Anti-Affinity)

Wykorzystano zaawansowane reguły harmonogramowania Kubernetes w celu optymalizacji wydajności sieciowej i odporności na awarie sprzętowe:

*   **Wysoka dostępność (Pod Anti-Affinity):**
    Wdrożona dla bezstanowych podów aplikacji w Go (np. `user-service`) z kluczem topologii `kubernetes.io/hostname`. Wymusza ona na planiście (Scheduler) rozłożenie replik tej samej usługi na **różnych węzłach fizycznych** (serwerach VM). W przypadku awarii jednego serwera Proxmox, druga replika wciąż działa na sprawnym węźle.
*   **Optymalizacja opóźnień (Pod Affinity):**
    Aplikacja Go intensywnie odpytuje swoją bazę danych PostgreSQL. Wdrożono regułę powinowactwa, która sugeruje planiście umieszczenie podu mikrousługi (np. `user-service`) na **tym samym fizycznym węźle**, na którym działa jej baza danych (np. `db-users-0`). Komunikacja odbywa się wtedy lokalnie (loopback/localhost), co eliminuje opóźnienia sieciowe związane z przesyłaniem pakietów między serwerami fizycznymi.


