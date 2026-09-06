# Security

Dieses Dokument beschreibt die gemeldeten Sicherheitseigenschaften des
Feature-Flag-Service, den Support- und Patch-Zeitraum, die Meldestelle für
Schwachstellen sowie die Software Bill of Materials (SBOM).

## Gemeldete Sicherheitseigenschaften

Der Dienst ist nach dem Grundsatz Security-by-Design aufgebaut und meldet die
folgenden Sicherheitseigenschaften:

- **Authentifizierung:** Verwaltungsendpunkte (`POST /flags`, `PUT /flags/{key}`,
  `DELETE /flags/{key}`, `GET /flags`) sind durch eine
  Authentifizierungs-/Autorisierungs-Middleware geschützt. Nicht authentifizierte
  Aufrufe werden mit einem JSON-Fehlerobjekt zurückgewiesen.
- **Rate-Limiting:** Eine Rate-Limit-Middleware begrenzt die Aufrufhäufigkeit
  pro Client und schützt den Dienst vor Überlastung und DoS.
- **Body-Limit:** Request-Bodies sind auf maximal 1 MiB begrenzt; größere
  Bodies werden vor dem vollständigen Einlesen mit `400` abgelehnt.
- **Timeouts:** Der `http.Server` ist mit `ReadHeaderTimeout`, `ReadTimeout`,
  `WriteTimeout` und `IdleTimeout` konfiguriert und schützt so vor
  Ressourcenbindung durch offene oder langsame Verbindungen (z. B. Slowloris).
- **Generische Fehlerantworten:** Alle `500`-Antworten enthalten im Feld
  `error` ausschließlich eine generische Meldung (z. B. `internal error`) ohne
  Stacktrace, Dateipfade oder interne Fehlerdetails. Eine Panic wird von der
  Recover-Middleware abgefangen und als `500`-JSON-Fehlerobjekt beantwortet,
  ohne den Serverprozess zu beenden.
- **Bindung an Loopback:** Der Dienst bindet standardmäßig nur an `127.0.0.1`.
  Die TLS-Terminierung erfolgt an einem vorgelagerten Reverse-Proxy.
- **Keine Speicherung personenbezogener Daten:** Nutzer-IDs werden ausschließlich
  transient zur deterministischen Evaluierung verarbeitet und weder gespeichert
  noch geloggt.

## Support- und Patch-Zeitraum

Der Dienst wird aktiv gewartet. Sicherheitsrelevante Korrekturen werden für den
aktuellen Haupt-Branch (und damit die aktuelle Release-Version) umgehend
bereitgestellt. Für zurückliegende Versionen besteht kein eigener
Support-Zeitraum; ein Update auf die aktuelle Version wird empfohlen, sobald
eine Schwachstelle behoben oder eine Änderung veröffentlicht wurde.

## Sicherheitsupdates einspielen

Neue Versionen werden als Build-Artefakt aus dem Repository erzeugt und der
laufende Prozess ersetzt (Details siehe `README.md`, Abschnitt „Betrieb &
Updates"). Sicherheitsrelevante Updates folgen demselben Weg und sollten
priorisiert eingespielt werden. Da der Dienst keinen persistenten Zustand hält,
ist dabei kein Datenübernahme-Schritt erforderlich.

## Meldestelle für Schwachstellen

Sicherheitslücken oder Schwachstellen können vertraulich per E-Mail an den
Betreiber des jeweiligen Deployments gemeldet werden. Bitte nennen Sie in der
Meldung:

- die betroffene Version,
- eine Beschreibung der Schwachstelle,
- Schritte zur Reproduktion (falls vorhanden),
- die erwarteten und die tatsächlichen Auswirkungen.

Sofern im Repository eine Betreiber-Adresse hinterlegt ist, ersetzt diese die
generische Kontaktaufnahme über den Projekt-Admin. Sicherheitsmeldungen werden
vertraulich behandelt; eine Veröffentlichung erfolgt erst nach Bereitstellung
einer Korrektur.

## Software Bill of Materials (SBOM)

Der Dienst verwendet ausschließlich die Go-Standardbibliothek und keine
Drittanbieter-Abhängigkeiten. Die SBOM ergibt sich aus der Moduldeklaration:

```
$ go list -m all
featureflags
```

- **Modul:** `featureflags`
- **Go-Version:** 1.22
- **Externe Abhängigkeiten:** keine (nur Standardbibliothek)

Damit bestehen keine verwundbaren oder nachzupflegenden Drittanbieter-Bibliotheken;
das Angriffsoberfläche der Abhängigkeiten beschränkt sich auf die von der
Go-Laufzeitumgebung bereitgestellte Standardbibliothek.
