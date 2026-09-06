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

## Datenschutz

Der `user`-Parameter des Evaluate-Endpunkts (`GET /flags/{key}/evaluate?user={id}`)
wird ausschließlich transient zur deterministischen Evaluierung verarbeitet:
Er fließt in den stabilen Hash (FNV-1a) ein und wird danach verworfen. Der
Dienst speichert keinerlei Nutzer-IDs — weder dauerhaft im In-Memory-Store noch
in Logs (die Logging-Middleware protokolliert nur Methode, Pfad, Statuscode und
Dauer, niemals den `user`-Wert). Verarbeitet werden ausschließlich
Flag-Daten (`key`, `enabled`, `description`, `rollout_percent`).

Die Rechtsgrundlage der Verarbeitung richtet sich nach dem Einsatzszenario des
Betreibers: bei einer Verarbeitung im Rahmen eines Vertragsverhältnisses
Art. 6 Abs. 1 lit. b DSGVO, andernfalls Art. 6 Abs. 1 lit. f DSGVO
(berechtigtes Interesse an der deterministischen Steuerung der
Feature-Auslieferung). Der Betreiber ist verantwortlich dafür, die konkrete
Rechtsgrundlage zu bestimmen und eine Datenschutzerklärung bereitzustellen, die
über Art, Umfang und Zweck der Verarbeitung informiert.

## Transport & TLS

Der Dienst bindet standardmäßig nur an `127.0.0.1` (Loopback). Für den Betrieb
jenseits der Maschinengrenze muss die TLS-Terminierung an einem vorgelagerten
Reverse-Proxy erfolgen; der Dienst selbst spricht kein TLS. Der `user`-Parameter
ist ein personenbezogenes Datum und darf ausschließlich verschlüsselt (über
TLS/HTTPS) übertragen werden. Ein direktes unverschlüsseltes Exponieren des
Dienstes ins Internet ist nicht vorgesehen und darf nicht erfolgen.

## Betrieb & Updates

Der Dienst ist ein eigenständiger Go-Prozess ohne externe Laufzeitabhängigkeiten.
Updates werden eingespielt, indem eine neue Build-Artefakt aus dem Repository
erzeugt und der laufende Prozess ersetzt wird:

```sh
go build ./...          # neues Artefakt bauen
```

Anschließend wird der bestehende Prozess gestoppt und der neu gebaute gestartet
(`go run .` in der Entwicklung bzw. das gebaute Binary im produktiven Betrieb).
Da der Dienst keinerlei persistenten Zustand hält (alle Flag-Daten liegen nur im
Arbeitsspeicher), ist nach einem Neustart kein Migrations- oder Datenübernahme-
Schritt erforderlich; der Store beginnt leer. Ein kontrolliertes Deployment
(Stopp → Build → Start) stellt sicher, dass währenddessen keine parallelen
Zugriffe auf einen veralteten Prozess erfolgen.
