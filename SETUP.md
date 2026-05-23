# Podręcznik instalacji i uruchomienia aplikacji w Minikube i K3s

Ten przewodnik krok po kroku opisuje sposób konfigurowania, wdrażania oraz testowania rozproszonego systemu rezerwacji kinowych **Cinema System** w środowisku Kubernetes (lokalny **Minikube** lub klaster **K3s** na Proxmox) z wykorzystaniem obrazów w wersji `1.0.4`.

---

## 1. Wymagania wstępne

Przed przystąpieniem do instalacji upewnij się, że na systemie lokalnym zainstalowane są następujące narzędzia:
- **Minikube** (v1.30.0 lub nowszy) lub działający klaster **K3s/Kubernetes**
- **kubectl** (skonfigurowany do obsługi klastra)
- **Docker** (opcjonalnie, do lokalnych testów)
- **curl** (do weryfikacji API)

---

## 2. Inicjalizacja klastra (opcjonalnie dla Minikube)

Jeśli wdrażasz aplikację lokalnie na Minikube, uruchom klaster przy użyciu sterownika Docker:

```bash
minikube start --driver=docker
```

Upewnij się, że klaster działa poprawnie:

```bash
minikube status
kubectl get nodes
```

---

## 3. Konfiguracja Wolumenów Pamięci i Uprawnień

Nasza architektura korzysta z wolumenów typu `PersistentVolume` z mapowaniem ścieżek fizycznych (`local-storage` na węźle klastra). Pliki PV są zabezpieczone klauzulą **`claimRef`**, co gwarantuje, że dany wolumen powiąże się **wyłącznie** ze swoją dedykowaną bazą danych (zapobiega to losowemu mieszaniu baz).

Ponieważ kontenery bazodanowe są utwardzone i działają jako użytkownik nieuprzywilejowany (`runAsUser: 999`), a **PostgreSQL wymaga rygorystycznych zabezpieczeń katalogu danych (uprawnienia 700)**, musisz utworzyć katalogi i nadać im odpowiedniego właściciela oraz uprawnienia na węźle, na którym działa baza danych.

### Opcja A: Wdrożenie na Minikube (Lokalnie)
Wykonaj poniższe polecenie SSH bezpośrednio na maszynie wirtualnej Minikube:
```bash
minikube ssh "sudo mkdir -p /mnt/data/db-users /mnt/data/db-catalog /mnt/data/db-showtime /mnt/data/db-booking && sudo chown -R 999:999 /mnt/data/db-* && sudo chmod 700 /mnt/data/db-*"
```

### Opcja B: Wdrożenie na klastrze K3s / Proxmox (Wielowęzłowym)
Zaloguj się na węzeł docelowy (np. `node-1`, na który wskazuje `nodeAffinity` w plikach PV) i wykonaj:
```bash
sudo mkdir -p /mnt/data/db-users /mnt/data/db-catalog /mnt/data/db-showtime /mnt/data/db-booking
sudo chown -R 999:999 /mnt/data/db-*
sudo chmod 700 /mnt/data/db-*
```

> [!IMPORTANT]
> Pominięcie ustawienia właściciela `999:999` lub nadanie zbyt otwartych uprawnień (np. `777`) spowoduje, że silnik PostgreSQL odmówi uruchomienia i pod wejdzie w stan `CrashLoopBackOff`.

---

## 4. Przygotowanie obrazów kontenerów z Docker Hub

System korzysta z gotowych obrazów umieszczonych w rejestrze **Docker Hub** w repozytorium użytkownika `adamad7` w wersji **`1.0.4`**.

Manifesty wdrożeniowe są domyślnie skonfigurowane tak, aby automatycznie pobierać poniższe obrazy z Docker Hub:
- Serwis użytkowników: `adamad7/user-service:1.0.4`
- Serwis katalogu: `adamad7/catalog-service:1.0.4`
- Serwis seansów: `adamad7/showtime-service:1.0.4`
- Serwis rezerwacji: `adamad7/booking-service:1.0.4`
- Aplikacja frontendowa: `adamad7/frontend:1.0.4`

### Opcjonalne: Ręczne pobranie obrazów do pamięci Minikube
```bash
minikube image pull adamad7/user-service:1.0.4
minikube image pull adamad7/catalog-service:1.0.4
minikube image pull adamad7/showtime-service:1.0.4
minikube image pull adamad7/booking-service:1.0.4
minikube image pull adamad7/frontend:1.0.4
```

---

## 5. Wdrożenie manifestów Kubernetes (Kolejność ma znaczenie!)

Wdrożenie komponentów musi przebiegać sekwencyjnie. Nasze aplikacje mają zdefiniowane **jawne limity zasobów dla kontenerów startowych `db-migration`**, co gwarantuje pełną zgodność z rygorystycznymi politykami `ResourceQuota` (np. na Minikube z wyłączonym kontrolerem `LimitRange`).

Uruchom poniższe polecenia po kolei:

```bash
# 1. Tworzenie przestrzeni nazw i limitów zasobów (Quota, LimitRange)
kubectl apply -f k8s/namespaces/

# 2. Tworzenie wolumenów (StorageClass i zablokowane za pomocą claimRef PersistentVolumes)
kubectl apply -f k8s/storage/

# 3. Sekrety (hasła DB, klucz JWT)
kubectl apply -f k8s/secrets/

# 4. Migracje bazodanowe (skrypty SQL w ConfigMaps)
kubectl apply -f k8s/configmaps/

# 5. Reguły sieciowe (Polityki sieciowe w tym zaktualizowany allow-backends wspierający Ingress)
kubectl apply -f k8s/networkpolicies/

# 6. Bazy danych (StatefulSets i usługi ClusterIP)
kubectl apply -f k8s/db-users/
kubectl apply -f k8s/db-catalog/
kubectl apply -f k8s/db-showtime/
kubectl apply -f k8s/db-booking/
```

> [!TIP]
> Poczekaj, aż pody baz danych osiągną status `READY 1/1` (`kubectl get pods -n backend-ns -w`). Gdy bazy będą gotowe, wdróż aplikacje backendowe, które automatycznie przeprowadzą migracje schematów SQL:

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

## 6. Weryfikacja uruchomienia i Dostęp do Aplikacji

Sprawdź stan podów we wszystkich przestrzeniach nazw:
```bash
kubectl get pods -A
```
Wszystkie pody powinny mieć status `Running` i stan `Ready` (np. `1/1` lub `2/2`).

### Dostęp przez Ingress na Minikube (Lokalnie)
Nasz Ingress nie posiada ograniczeń domenowych (`host`), co pozwala na bezpośrednie zapytania pod adres IP klastra:
1. Pobierz IP Minikube:
   ```bash
   minikube ip
   ```
2. Otwórz przeglądarkę i wejdź na adres:
   ```text
   http://<ADRES_IP_MINIKUBE>/
   ```

### Dostęp na klastrze K3s / Proxmox
Dostęp odbywa się poprzez wystawiony kontroler Ingress i skonfigurowany tunel Cloudflare pod Twoją domeną (np. `https://cinema.mazadonia.cc/`). Dzięki regułom w `allow-backends.yaml`, kontroler Ingress działający w `ingress-nginx`/`kube-system` bez problemu prześle zapytania `/api/...` bezpośrednio do usług backendowych (brak błędu 504 Gateway Timeout).

---

## 7. Scenariusze testowe i polecenia weryfikacji API

Skonfiguruj zmienną środowiskową `APP_URL` wskazującą na adres aplikacji (np. `http://192.168.49.2` dla Minikube lub `https://cinema.mazadonia.cc` dla klastra produkcyjnego):

```bash
export APP_URL="http://192.168.49.2" # zamień na swój adres
```

### Scenariusz 1: Pobranie listy filmów (Catalog Service)
- **Zapytanie HTTP**: `GET /api/movies`
- **Polecenie**:
  ```bash
  curl -i $APP_URL/api/movies
  ```
- **Oczekiwany rezultat**: Kod `200 OK` i lista filmów w formacie JSON (np. *The Matrix*, *Inception*).

### Scenariusz 2: Rejestracja nowego użytkownika (User Service)
- **Zapytanie HTTP**: `POST /api/users/register`
- **Polecenie**:
  ```bash
  curl -i -X POST -H "Content-Type: application/json" \
    -d '{"username":"jan_kowalski", "email":"jan@kowalski.pl", "password":"BezpieczneHaslo1!"}' \
    $APP_URL/api/users/register
  ```
- **Oczekiwany rezultat**: Kod `201 Created` i dane użytkownika z hashem hasła.

### Scenariusz 3: Test Izolacji Sieciowej (NetworkPolicy)
Sprawdzenie działania rygorystycznych reguł Cilium – pod w przestrzeni `frontend-ns` nie może łączyć się bezpośrednio z bazą danych w `backend-ns` (cały ruch musi przechodzić przez usługi backendu):

```bash
kubectl run network-test --rm -i --tty --image=alpine --namespace=frontend-ns --sh -c "apk add --no-cache postgresql-client && pg_isready -h db-users.backend-ns.svc.cluster.local -p 5432"
```
- **Oczekiwany rezultat**: Połączenie zostanie całkowicie zablokowane na warstwie sieciowej przez eBPF Cilium (brak odpowiedzi / timeout).
