# Cinema System - Mikroserwisy, Docker i Docker Compose

Projekt prezentuje system rezerwacji kinowych oparty o mikroserwisy:

- `user-service`
- `catalog-service`
- `showtime-service`
- `booking-service`
- `frontend`

Całość uruchamiana jest przez Docker Compose.

## Architektura

- Baza danych PostgreSQL per domena (`users`, `catalog`, `showtime`, `booking`)
- Osobne serwisy backendowe w Go
- Frontend React + Vite serwowany przez Nginx
- Migracje baz uruchamiane jako dedykowane kontenery

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
