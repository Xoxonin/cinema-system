# Podręcznik instalacji i uruchomienia aplikacji w Minikube (SETUP.md)

Ten przewodnik krok po kroku opisuje sposób konfigurowania, wdrażania oraz testowania rozproszonego systemu rezerwacji kinowych (**Cinema System**) w środowisku Kubernetes za pomocą narzędzia **Minikube** z wykorzystaniem gotowych obrazów z rejestru **Docker Hub** (wersja `1.0.3`).

---

## 1. Wymagania wstępne (Prerequisites)

Przed przystąpieniem do instalacji upewnij się, że na systemie lokalnym zainstalowane są następujące narzędzia:
- **Minikube** (v1.30.0 lub nowszy)
- **kubectl** (skonfigurowany do obsługi klastra)
- **Docker** (zainstalowany i uruchomiony lokalnie)
- **curl** (do weryfikacji API)

---

## 2. Inicjalizacja klastra Minikube

Uruchom lokalny klaster Kubernetes przy użyciu sterownika Docker:

```bash
minikube start --driver=docker
```

Upewnij się, że klaster działa poprawnie i węzeł jest gotowy:

```bash
minikube status
kubectl get nodes
```

---

## 3. Konfiguracja Wolumenów Pamięci

Nasza architektura korzysta z wolumenów typu `PersistentVolume` z mapowaniem ścieżek fizycznych (`local-storage` na węźle klastra). 

Ponieważ kontenery bazodanowe są utwardzone i działają jako użytkownik nieuprzywilejowany (`runAsUser: 999`), należy utworzyć odpowiednie katalogi oraz nadać im wymagane uprawnienia bezpośrednio na maszynie wirtualnej Minikube:

```bash
# Połączenie SSH z minikube w celu utworzenia katalogów i nadania uprawnień właściciela użytkownikowi o ID 999 (postgres)
minikube ssh "sudo mkdir -p /mnt/data/db-users /mnt/data/db-catalog /mnt/data/db-showtime /mnt/data/db-booking && sudo chmod 777 /mnt/data/db-* && sudo chown -R 999:999 /mnt/data/db-*"
```

---

## 4. Przygotowanie obrazów kontenerów z Docker Hub

System jest w pełni przystosowany do uruchomienia przy użyciu gotowych, produkcyjnych obrazów umieszczonych w rejestrze **Docker Hub** w repozytorium użytkownika `adamad7` w wersji `1.0.3`.

Dzięki temu **nie ma potrzeby lokalnego budowania obrazów** ani konfigurowania lokalnego środowiska Docker na Minikube. Manifesty wdrożeniowe Kubernetes są domyślnie skonfigurowane tak, aby automatycznie pobierać poniższe obrazy z Docker Hub:

- Serwis użytkowników: `adamad7/user-service:1.0.3`
- Serwis katalogu: `adamad7/catalog-service:1.0.3`
- Serwis seansów: `adamad7/showtime-service:1.0.3`
- Serwis rezerwacji: `adamad7/booking-service:1.0.3`
- Aplikacja frontendowa: `adamad7/frontend:1.0.3`

### Opcjonalne: Ręczne pobranie obrazów do pamięci klastra
Jeśli chcesz pobrać obrazy wcześniej (np. w celu skrócenia czasu pierwszego uruchomienia wdrożenia lub pracy offline), możesz wydać polecenia pobrania obrazów bezpośrednio do pamięci klastra Minikube:

```bash
minikube image pull adamad7/user-service:1.0.3
minikube image pull adamad7/catalog-service:1.0.3
minikube image pull adamad7/showtime-service:1.0.3
minikube image pull adamad7/booking-service:1.0.3
minikube image pull adamad7/frontend:1.0.3
```

Możesz zweryfikować poprawność pobranych obrazów wewnątrz Minikube za pomocą:
```bash
minikube image ls | grep adamad7
```

---

## 5. Wdrożenie manifestów Kubernetes (Precise Order)

Wdrażanie komponentów musi przebiegać w określonej kolejności, aby zapewnić istnienie zależności (przestrzenie nazw, wolumeny, sekrety, bazy danych, a dopiero potem aplikacje i Ingress).

Uruchom poniższe polecenia sekwencyjnie:

```bash
# 1. Tworzenie przestrzeni nazw i limitów (Quota, LimitRange)
kubectl apply -f k8s/namespaces/

# 2. Tworzenie wolumenów (StorageClass i PersistentVolume)
kubectl apply -f k8s/storage/

# 3. Sekrety (db passwords, JWT secret)
kubectl apply -f k8s/secrets/

# 4. Migracje bazodanowe (skrypty SQL w ConfigMaps)
kubectl apply -f k8s/configmaps/

# 5. Reguły sieciowe (NetworkPolicies)
kubectl apply -f k8s/networkpolicies/

# 6. Bazy danych (StatefulSets i usługi ClusterIP)
kubectl apply -f k8s/db-users/
kubectl apply -f k8s/db-catalog/
kubectl apply -f k8s/db-showtime/
kubectl apply -f k8s/db-booking/
```

> [!TIP]
> Przed wdrożeniem aplikacji backendowych warto upewnić się, że pody bazodanowe w pełni wystartowały (status `READY 1/1`), ponieważ initContainers w aplikacjach wykonają automatyczne migracje schematów SQL.
> Monitoruj stan baz danych za pomocą: `kubectl get pods -n backend-ns -w`

Po pomyślnym uruchomieniu baz danych zaaplikuj pozostałe komponenty:

```bash
# 7. Aplikacje backendowe (Deployments i usługi ClusterIP)
kubectl apply -f k8s/user-service/
kubectl apply -f k8s/catalog-service/
kubectl apply -f k8s/showtime-service/
kubectl apply -f k8s/booking-service/

# 8. Frontend i Ingress
kubectl apply -f k8s/frontend/
kubectl apply -f k8s/ingress/
```

---

## 6. Weryfikacja uruchomienia aplikacji

Sprawdź stan wszystkich podów w obu przestrzeniach nazw:
```bash
kubectl get pods -A
```
Prawidłowy stan to `READY 1/1` i `STATUS Running` dla wszystkich elementów.

Pobierz adres URL, pod którym aplikacja jest dostępna przez NodePort:
```bash
minikube service frontend-nodeport -n frontend-ns --url
```
Zwrócony adres (np. `http://192.168.49.2:30080`) posłuży do testowania aplikacji. Zdefiniujmy go w zmiennej środowiskowej dla ułatwienia testów (w powłoce Bash):
```bash
export APP_URL=$(minikube service frontend-nodeport -n frontend-ns --url)
```

---

## 7. Scenariusze testowe i polecenia weryfikacji API

Poniższe scenariusze odwzorowują interakcję z systemem za pośrednictwem bezpiecznej bramy (Nginx proxy) na porcie frontendu.

### Scenariusz 1: Pobranie listy filmów (Catalog Service)
Zweryfikowanie połączenia bramy z serwisem katalogu oraz automatycznego zasilenia bazy danych filmami.
- **Zapytanie HTTP**: `GET /api/movies`
- **Polecenie testowe**:
```bash
curl -i $APP_URL/api/movies
```
- **Oczekiwany rezultat**: Kod `200 OK` i lista filmów w formacie JSON (zawierająca m.in. *The Matrix*, *Inception*, *Dune: Part Two*):
```json
[{"id":1,"title":"The Matrix","description":{"String":"...","Valid":true},"duration_minutes":136,...}]
```

---

### Scenariusz 2: Rejestracja nowego użytkownika (User Service)
Rejestracja konta użytkownika z szyfrowaniem hasła algorytmem bcrypt.
- **Zapytanie HTTP**: `POST /api/users/register`
- **Polecenie testowe**:
```bash
curl -i -X POST -H "Content-Type: application/json" \
  -d '{"username":"jan_kowalski", "email":"jan@kowalski.pl", "password":"BezpieczneHaslo1!"}' \
  $APP_URL/api/users/register
```
- **Oczekiwany rezultat**: Kod `201 Created` i dane nowo zarejestrowanego użytkownika z ukryciem hasła w postaci hashu:
```json
{"id":1,"username":"jan_kowalski","email":"jan@kowalski.pl","role":"user","created_at":...}
```

---

### Scenariusz 3: Uwierzytelnienie i logowanie użytkownika (User Service)
Generowanie bezpiecznego tokenu JWT.
- **Zapytanie HTTP**: `POST /api/users/login`
- **Polecenie testowe**:
```bash
curl -i -X POST -H "Content-Type: application/json" \
  -d '{"email":"jan@kowalski.pl", "password":"BezpieczneHaslo1!"}' \
  $APP_URL/api/users/login
```
- **Oczekiwany rezultat**: Kod `200 OK` i token autoryzacyjny:
```json
{"token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."}
```

> [!TIP]
> Zapisz wygenerowany token (bez cudzysłowów) w zmiennej dla kolejnego testu:
> `export TOKEN="twoj_token_jwt"`

---

### Scenariusz 4: Pobranie profilu zalogowanego użytkownika (User Service)
Autoryzacja zapytania z użyciem tokena JWT w nagłówku.
- **Zapytanie HTTP**: `GET /api/users/profile`
- **Polecenie testowe**:
```bash
curl -i -H "Authorization: Bearer $TOKEN" $APP_URL/api/users/profile
```
- **Oczekiwany rezultat**: Kod `200 OK` z profilem użytkownika powiązanym z tokenem:
```json
{"email":"jan@kowalski.pl","id":1,"role":"user","username":"jan_kowalski"}
```

Jeśli wyślesz zapytanie bez nagłówka lub z błędnym tokenem:
```bash
curl -i $APP_URL/api/users/profile
```
Otrzymasz poprawną odmowę dostępu: `401 Unauthorized`.

---

### Scenariusz 5: Odpytanie o pokoje kinowe i seanse (Showtime Service)
Pobranie automatycznie wygenerowanej mapy sal kinowych i harmonogramu seansów.
- **Zapytanie o pokoje**: `GET /api/rooms`
- **Polecenie testowe**:
```bash
curl -i $APP_URL/api/rooms
```
- **Zapytanie o seanse dla filmu o ID = 1**: `GET /api/showtimes?movie_id=1`
- **Polecenie testowe**:
```bash
curl -i "$APP_URL/api/showtimes?movie_id=1"
```
- **Oczekiwany rezultat**: Kody `200 OK` wraz z listą dostępnych sal oraz seansów dopasowanych do wybranego filmu.

---

### Scenariusz 6: Test Izolacji Sieciowej (NetworkPolicy)
Sprawdzenie, czy reguły sieciowe (deny-by-default) skutecznie blokują ruch nieautoryzowany. Zgodnie z architekturą, pody w przestrzeni `frontend-ns` NIE MOGĄ łączyć się bezpośrednio z bazami danych w przestrzeni `backend-ns` (dostęp ma tylko backend).

**Polecenie testowe:**
Uruchom pod w przestrzeni frontendu i spróbuj połączyć się z bazą danych na porcie 5432:
```bash
kubectl run network-test --rm -i --tty --image=alpine --namespace=frontend-ns --sh -c "apk add --no-cache postgresql-client && pg_isready -h db-users.backend-ns.svc.cluster.local -p 5432"
```

- **Oczekiwany rezultat**: Połączenie zostanie zablokowane i upłynie limit czasu oczekiwania (brak odpowiedzi), ponieważ zasady NetworkPolicy filtrują ten nieautoryzowany ruch na poziomie L4.
```
db-users.backend-ns.svc.cluster.local:5432 - no response
```
