# Feature-Flag-Service

Ein eigenständiger REST-Dienst in Go, der Feature-Flags in einem
thread-sicheren In-Memory-Store verwaltet. Er legt Flags an, listet sie auf,
aktualisiert und löscht sie und liefert über einen Endpunkt deterministische
Ja/Nein-Entscheidungen pro Nutzer (Rollout).

## Tech-Stack

- Go 1.22 (nur Standardbibliothek, keine externen Abhängigkeiten)
- Routing: `net/http` mit Go-1.22-Method-Mustern
- JSON: `encoding/json`
- Datenhaltung: In-Memory-Store, geschützt durch `sync.RWMutex`
- Logging: `log` aus der Standardbibliothek als Middleware
- Tests: Go-Standard `testing` mit `net/http/httptest`

## Installation

Voraussetzung ist eine Go-Installation ab Version 1.22. Es sind keine
externen Abhängigkeiten zu installieren.

## Ausführen

```sh
go run .
```

Der Dienst startet auf Port `8080`.

## Build

```sh
go build ./...
```

## Tests

```sh
go test ./...
```

## Endpunkte

Alle Endpunkte liefern `Content-Type: application/json`. Fehlerobjekte haben
immer die Form `{"error":"…"}`.

| Methode | Pfad                          | Beschreibung                                    |
|---------|-------------------------------|-------------------------------------------------|
| GET     | `/healthz`                    | Health-Check, antwortet `200` mit `{"status":"ok"}` |
| POST    | `/flags`                      | Legt ein Flag an (`201`), `400` bei ungültigen Werten, `409` bei vorhandenem key |
| GET     | `/flags`                      | Liefert alle Flags (`200`), leere Liste `[]` wenn keine existieren |
| GET     | `/flags/{key}`                | Liefert ein Flag (`200`), `404` bei unbekanntem key |
| PUT     | `/flags/{key}`                | Aktualisiert ein Flag (`200`), `400`/`404` |
| DELETE  | `/flags/{key}`                | Löscht ein Flag (`204`, leerer Body), `404` |
| GET     | `/flags/{key}/evaluate?user={id}` | Deterministische Ja/Nein-Entscheidung (`200` mit `{"result":bool}`) |

Ein Flag hat die Form:

```json
{"key":"string","enabled":true,"description":"string","rollout_percent":100}
```

## Features

- Thread-sicherer In-Memory-Store mit `sync.RWMutex`
- Begrenzung auf maximal 10 000 gleichzeitig gespeicherte Flags
- Deterministische Rollout-Entscheidung über einen stabilen Hash (FNV-1a)
- Logging-Middleware (Methode, Pfad, Statuscode, Dauer — ohne `user`-Werte)
- Panic-Recovery- und Body-Limit-Middleware (max. 1 MiB)
- Keine Speicherung von `user`-Werten — der Store hält ausschließlich
  Flag-Daten (`key`, `enabled`, `description`, `rollout_percent`)
