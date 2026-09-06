VERDICT: BLOCKED

## Security-Report

**Hinweis:** Es wurde kein verwertbarer Scanner-Output mitgeliefert („no applicable security scanners for this project type“). Die Bewertung basiert daher auf manueller Codeanalyse.

---

### [HIGH] Fehlende Authentifizierung und Autorisierung für alle Endpunkte

**Betroffene Stellen:** `main.go` (Routing), `internal/httpapi/create.go`, `internal/httpapi/update.go`, `internal/httpapi/read.go`, `internal/httpapi/evaluate.go`

**Beschreibung:**  
Der Dienst lauscht ohne jede Authentifizierung auf Port 8080. Jeder Netzwerkteilnehmer, der den Port erreicht, kann Feature-Flags anlegen, ändern und löschen. Zusätzlich stammt die Nutzer-ID bei `GET /flags/{key}/evaluate?user={id}` direkt aus dem Query-Parameter; da der Aufrufer nicht verifiziert wird, kann er beliebige `user`-Werte verwenden und so die Rollout-Entscheidung gezielt beeinflussen (Bruteforce oder Offline-Berechnung des FNV-Hashes).

**Risiko:** Unautorisierte Manipulation der Feature-Auslieferung, Umgehen von Rollout-Logik, DoS durch Vollschreiben des Stores.

**Fix:**
- Einführung einer Authentifizierung, z. B. API-Key, OAuth2/JWT oder mTLS.
- Schreiboperationen (`POST`, `PUT`, `DELETE`) nur für autorisierte Clients zulassen.
- Die Nutzer-ID aus einem verifizierten Token ableiten und nicht aus dem Query-Parameter übernehmen.
- Falls der Dienst ausschließlich lokal verwendet wird: an `127.0.0.1` binden statt an alle Interfaces.

---

### [MEDIUM] Unbegrenzte Feldlängen ermöglichen Speicher-DoS

**Betroffene Stellen:** `internal/httpapi/create.go`, `internal/httpapi/update.go`, `internal/httpapi/validate.go`, `internal/store/store.go`

**Beschreibung:**  
Das Body-Limit von 1 MiB begrenzt einzelne Requests, aber `key` und `description` haben keine Längenbegrenzung. Bei `maxFlags = 10000` kann ein Angreifer mit Schreibzugriff theoretisch über 10 GiB an Strings im In-Memory-Store ablegen. Auch der `user`-Query-Parameter ist nur durch das HTTP-Header-Limit begrenzt.

**Risiko:** Speichererschöpfung und damit Denial of Service.

**Fix:**
- Maximale Längen definieren und in der Validierung erzwingen, z. B. `len(key) <= 128` und `len(description) <= 2048`.
- Länge des `user`-Parameters in `Evaluate` prüfen und zu lange Werte mit `400` ablehnen.
- Optional ein Gesamtspeicherlimit oder eine maximale kumulierte Flag-Größe einführen.

---

### [MEDIUM] Ungesicherter Transport und fehlende Server-Timeouts

**Betroffene Stelle:** `main.go` → `http.ListenAndServe(":8080", newHandler())`

**Beschreibung:**  
Der Server lauscht auf allen Interfaces, verwendet unverschlüsseltes HTTP und der Standard-`http.Server` hat keine Read-/Write-/Idle-Timeouts. Das erlaubt Netzwerk-Sniffing der Flag-Metadaten und macht den Dienst anfällig für langsame Request-Angriffe (Slowloris).

**Risiko:** Abhören von Flag-Informationen, Ressourcenbindung durch offene Verbindungen.

**Fix:**
- Expliziten `http.Server` mit Timeouts verwenden:
  ```go
  srv := &http.Server{
      Addr:              "127.0.0.1:8080", // oder internes Interface
      Handler:           newHandler(),
      ReadHeaderTimeout: 5 * time.Second,
      ReadTimeout:       10 * time.Second,
      WriteTimeout:      10 * time.Second,
      IdleTimeout:       60 * time.Second,
  }
  ```
- TLS terminieren oder den Dienst hinter einem TLS-Proxy betreiben, sofern er nicht nur auf Loopback läuft.

---

### [LOW] Panic-Requests werden nicht geloggt

**Betroffene Stelle:** `internal/httpapi/middleware.go`

**Beschreibung:**  
`Recover` liegt außerhalb von `Logging`. Wird eine Panic ausgelöst, fängt `Recover` sie ab, aber der Logging-Code nach `next.ServeHTTP` wird nicht mehr ausgeführt. Dadurch fehlen für diese Requests Methode, Pfad, Status und Dauer im Log – was die Fehlerdiagnose erschwert und AC-09 verletzt.

**Fix:**  
Reihenfolge oder Implementierung anpassen, z. B. `Logging` als äußerste Middleware verwenden oder in `Logging` ein `defer` einbauen, das den Request auch bei einer Panic protokolliert. Alternativ die Panic-Antwort in `Recover` über den `statusWriter` laufen lassen.

---

### [LOW] JSON-Decoder zu tolerant

**Betroffene Stellen:** `internal/httpapi/create.go`, `internal/httpapi/update.go`

**Beschreibung:**  
`json.NewDecoder(r.Body).Decode(&req)` ignoriert unbekannte Felder und akzeptiert nach dem ersten JSON-Objekt weitere Daten. Dadurch können vertippte Felder unbemerkt bleiben und unerwartete Payloads verarbeitet werden.

**Risiko:** Gering; erschwert robuste Eingabevalidierung und kann zu unerwartetem Verhalten führen.

**Fix:**
- `dec.DisallowUnknownFields()` verwenden.
- Nach dem `Decode` prüfen, ob nur ein JSON-Wert vorliegt, z. B. durch `dec.More()` bzw. einen zweiten `Decode`-Aufruf, der `io.EOF` liefern muss.
- Optional `Content-Type: application/json` prüfen.

---

### [LOW] Deterministischer, nicht kryptographischer Hash in Kombination mit clientseitig wählbarem User

**Betroffene Stelle:** `internal/httpapi/hash.go`, `internal/httpapi/evaluate.go`

**Beschreibung:**  
FNV-1a ist deterministisch und öffentlich bekannt. Solange die `user`-ID vom Client gewählt wird, kann ein Angreifer Offline berechnen oder durchprobieren, welche `user`-Werte bei einem Flag `true` liefern. Das ist kein isolierter Fehler des Hash-Algorithmus – die Spezifikation verlangt deterministische Entscheidungen –, wird aber durch die fehlende Authentifizierung zum Manipulationsvektor.

**Fix:**  
Primär durch die oben beschriebene Authentifizierung und serverseitige Bindung der Nutzeridentität beheben. Falls eine erhöhte Undurchschaubarkeit gewünscht ist, kann ein HMAC mit einem serverseitigen Secret verwendet werden; die deterministische Natur bleibt dabei erhalten.

---

**Fazit:**  
Die fehlende Zugriffskontrolle für sämtliche verwaltenden Endpunkte ist ein hohes Sicherheitsrisiko und macht den Dienst in der vorliegenden Form nicht freigabefähig. Die übrigen Punkte sind Härtungsmaßnahmen, die vor oder mit der Freigabe umgesetzt werden sollten.