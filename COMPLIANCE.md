VERDICT: BLOCKED

# Compliance-Report: Feature-Flag-Service (Go-Backend)

**Projekttyp:** `go-backend` — reine REST-API ohne Endnutzer-UI. UI-/Cookie-/Accessibility-Pflichten sind daher nicht einschlägig; bewertet werden DSGVO und CRA sowie der Sicherheitszustand des sichtbaren Codes.

---

## 1. DSGVO (Datenschutz)

### 1.1 Kritisch: Personenbezogene Daten werden unverschlüsselt über HTTP übertragen
**Betroffene Datei:** `main.go`

```go
if err := http.ListenAndServe(":8080", newHandler()); err != nil {
```

Der Evaluierungsendpunkt `GET /flags/{key}/evaluate?user={id}` verarbeitet einen `user`-Parameter. Nutzer-IDs sind personenbezogene Daten (mindestens pseudonym, Art. 4 Nr. 1 DSGVO). Der Server bindet ausschließlich unverschlüsselt auf Port `8080`; es gibt keine TLS-Konfiguration, keine `http.Server`-Absicherung und keinen sichtbaren Zwang zur TLS-Terminierung an einem vorgelagerten Proxy. Damit ist die Vertraulichkeit der über den Query-String übertragenen Nutzer-IDs standardmäßig nicht gewährleistet (Art. 5 Abs. 1 lit. f, Art. 32 DSGVO).

**Abhilfe:**
- In `main.go` einen `http.Server` mit `ListenAndServeTLS` verwenden und Zertifikatspfade über Umgebungsvariablen konfigurierbar machen, **oder**
- im Deployment verbindlich einen TLS-terminierenden Reverse-Proxy vorschreiben und im Code sicherstellen, dass der Dienst nur an ein internes Interface bindet (z. B. `127.0.0.1:8080`, konfigurierbar über Environment).
- Zusätzlich in der Dokumentation (`README.md`) festhalten, dass der `user`-Parameter personenbezogen ist und der Transport vertraulich erfolgen muss.

### 1.2 Hoch: Keine Authentifizierung/Autorisierung für Verwaltungsendpunkte
**Betroffene Dateien:** `main.go`, `internal/httpapi/create.go`, `internal/httpapi/update.go`, `internal/httpapi/read.go`

Die Endpunkte `POST /flags`, `PUT /flags/{key}`, `DELETE /flags/{key}` und `GET /flags` sind vollständig ungeschützt. Jeder Netzwerkteilnehmer kann Flags anlegen, ändern, löschen und auslesen. Das verletzt die Integrität und Vertraulichkeit der verarbeiteten Daten und stellt einen Verstoß gegen die technischen und organisatorischen Maßnahmen nach Art. 32 DSGVO dar. Es ermöglicht zudem Missbrauch des Services und Manipulation der Evaluierungsergebnisse.

**Abhilfe:**
- In `main.go` eine Authentifizierungs-/Autorisierungs-Middleware ergänzen (z. B. statischer Bearer-Token/API-Key, aus der Konfiguration gelesen, mit sicherem Vergleich).
- Für die Verwaltungsendpunkte (`POST`, `PUT`, `DELETE`, eventuell `GET /flags`) ausschließlich authentifizierte Aufrufe zulassen.
- Der Evaluierungsendpunkt `GET /flags/{key}/evaluate` kann bei fachlicher Entscheidung öffentlich bleiben; das muss in der Betriebsdokumentation begründet werden.

### 1.3 Niedrig: Panic-Logging kann unkontrolliert interne Details protokollieren
**Betroffene Datei:** `internal/httpapi/middleware.go`

```go
log.Printf("recovered from panic: %v", rec)
```

Der Wert der Panic (`rec`) kann interne Fehlerdetails, potenziell auch den Inhalt des aktuellen Requests enthalten. Die 500-Antwort ist generisch (siehe AC-13), aber das interne Log ist unkontrolliert.

**Abhilfe:**
- In `Recover` nur einen generischen Logeintrag schreiben, z. B.:
```go
log.Printf("recovered from panic")
```
- Oder nur den Panic-Typ (`fmt.Sprintf("%T", rec)`) protokollieren, ohne Werte.

### 1.4 Niedrig: Rechtsgrundlage und Datenschutzdokumentation fehlen
**Betroffene Datei:** `README.md`

Die Verarbeitung des `user`-Parameters ist im Code datenminimierend (keine Speicherung, keine Rückgabe, kein Logging des Query-Strings). Es fehlt aber eine Dokumentation der Rechtsgrundlage (z. B. Art. 6 Abs. 1 lit. b/f DSGVO), der Datenkategorien und der Tatsache, dass keine Speicherung erfolgt. Für einen reinen Backend-Dienst ist das keine UI-Pflicht, aber für den verantwortlichen Betreiber erforderlich.

**Abhilfe:**
- In `README.md` einen kurzen Abschnitt „Datenschutz/Datenschutzinformation“ ergänzen mit:
  - Verarbeitung von `user` ausschließlich transient zur deterministischen Evaluierung,
  - keine Speicherung von Nutzer-IDs,
  - Rechtsgrundlage gemäß Einsatzszenario,
  - Hinweis, dass der Betreiber eine Datenschutzerklärung bereitstellen muss.

---

## 2. EU Cyber Resilience Act (CRA)

### 2.1 Hoch: Keine dokumentierten Sicherheitseigenschaften, kein SBOM, kein Update-/Patch-Konzept
**Betroffene Datei:** `README.md`, `go.mod`

Der Dienst verwendet laut Code nur die Go-Standardbibliothek; eine SBOM ist damit überschaubar, aber nicht vorhanden. Es fehlen außerdem dokumentierte Sicherheitseigenschaften, ein Prozess für Sicherheitsupdates und ein Support-/Patch-Zeitraum. Für ein Produkt mit digitalen Elementen verlangt der CRA Security-by-Design, Transparenz und eine dokumentierte Schwachstellenbehandlung.

**Abhilfe:**
- Eine `SECURITY.md` mit gemeldeten Sicherheitseigenschaften, Update-Prozess, Supportzeitraum und Meldestelle für Schwachstellen ergänzen.
- Eine SBOM erzeugen und im Repository ablegen (z. B. `go version -m` / `go list -m all` als Basis, auch bei nur Standardbibliotheken).
- In `README.md` dokumentieren, wie Updates eingespielt werden (z. B. neues Container-Image, Deployment-Prozess).

### 2.2 Hoch: HTTP-Server ohne Timeouts und Härtung
**Betroffene Datei:** `main.go`

Die Verwendung von `http.ListenAndServe(":8080", ...)` setzt keine `ReadTimeout`, `WriteTimeout`, `IdleTimeout` oder `MaxHeaderBytes`. Das begünstigt Ressourcenerschöpfung (Slowloris) und entspricht nicht dem Security-by-Default-Gedanken des CRA.

**Abhilfe:**
- In `main.go` einen `http.Server` mit Timeouts konfigurieren:
```go
srv := &http.Server{
    Addr:              ":8080",
    Handler:           newHandler(),
    ReadHeaderTimeout: 5 * time.Second,
    ReadTimeout:       10 * time.Second,
    WriteTimeout:      10 * time.Second,
    IdleTimeout:       60 * time.Second,
    MaxHeaderBytes:    1 << 20,
}
```
- Den Server über `srv.ListenAndServeTLS` (oder TLS-Terminierung dokumentieren) starten.

### 2.3 Mittel: Keine Ratenbegrenzung/DoS-Schutz auf API-Ebene
**Betroffene Dateien:** `main.go`, `internal/httpapi/middleware.go`

Die API ist offen, und es existiert keine Begrenzung der Aufrufrate. Ein einzelner Client kann den Dienst mit Anfragen überlasten. Zwar schützt `LimitBody` vor großen Bodies, aber nicht vor hoher Request-Frequenz.

**Abhilfe:**
- Eine einfache Token-Bucket- oder Sliding-Window-Rate-Limit-Middleware pro Client-IP/Token in `internal/httpapi/middleware.go` implementieren und in `newHandler` in die Kette aufnehmen.
- Das Limit konfigurierbar machen (Umgebungsvariable).

### 2.4 Niedrig: Keine sicherheitsrelevanten Response-Header gegen Caching
**Betroffene Datei:** `internal/httpapi/respond.go`

Antworten enthalten kein `Cache-Control: no-store` und kein `X-Content-Type-Options: nosniff`. Bei einem Dienst, der personenbezogene Daten über URL transportiert, kann Caching zu ungewollter Offenlegung führen.

**Abhilfe:**
- In `writeJSON` zusätzlich setzen:
```go
w.Header().Set("Cache-Control", "no-store")
w.Header().Set("X-Content-Type-Options", "nosniff")
```

---

## 3. EU AI Act

**Keine KI-Funktion vorhanden.** Der Dienst implementiert deterministisches Feature-Flag-Rollout über FNV-1a-Hashing; es liegt kein KI-System im Sinne der Verordnung vor. Es ergeben sich keine AI-Act-Pflichten.

---

## 4. Pflichttexte & UI / Accessibility

**Nicht anwendbar.** Das Produkt ist ein reines Backend ohne Endnutzer-UI. Es gibt keine Cookies, kein Impressum/keine Legal-Notice-Pflicht im Code und keine Web-Oberfläche, die WCAG/BITV/EAA erfüllen müsste. Die datenschutzrechtliche Informationspflicht des Betreibers ist unter Ziffer 1.4 behandelt.

---

## 5. Zusammenfassung und Blocker-Begründung

Der Dienst erfüllt viele datenschutzfreundliche Anforderungen (kein Logging von Query-Strings, keine Speicherung von Nutzer-IDs, generische 500-Antworten, Body-Limit, Flag-Limit). Trotzdem bestehen zwei fundamentale Sicherheitsmängel:

1. **Unverschlüsselte Übertragung personenbezogener Daten** (`user`-Parameter im Klartext über HTTP ohne TLS-Konfiguration).
2. **Vollständig ungeschützte Verwaltungs-API** ohne Authentifizierung/Autorisierung.

Beide Mängel betreffen die Vertraulichkeit/Integrität der Verarbeitung und verstoßen gegen Art. 32 DSGVO; sie sind vor einer Marktfreigabe zwingend zu beheben.

**Empfohlene Priorität zur Behebung:**
1. TLS/Transportverschlüsselung oder verbindliche TLS-Terminierung (`main.go`).
2. Authentifizierung für `POST/PUT/DELETE /flags` und `GET /flags` (`main.go` + Middleware).
3. HTTP-Server-Timeouts und Härtung (`main.go`).
4. Security-Dokumentation, SBOM, Update-Konzept (`README.md`, `SECURITY.md`).
5. Response-Header gegen Caching (`internal/httpapi/respond.go`).