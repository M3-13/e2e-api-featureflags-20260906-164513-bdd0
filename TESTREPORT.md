VERDICT: BUGS_FOUND

**Titel:** Doppelt deklarierte Test-Hilfsfunktion `doRequest` verhindert das Kompilieren des HTTP-API-Testpakets

**Symptom:** `go test ./...` bricht mit Build-Fehler im Paket `featureflags/internal/httpapi` ab. Damit ist das Akzeptanzkriterium AC-10 (`go test ./...` läuft grün) verletzt und die automatisierte Verifikation des Feature-Flag-Service ist rot.

**Repro:** Im Projektstamm `go test ./...` ausführen.

**Evidence:**
```
FAIL	featureflags/internal/httpapi [build failed]
internal\httpapi\update_test.go:25:6: doRequest redeclared in this block
	internal\httpapi\flags_test.go:49:6: other declaration of doRequest
internal\httpapi\update_test.go:42:53: not enough arguments in call to doRequest
	have (*http.ServeMux, string, string, string)
	want (*testing.T, http.Handler, string, string, string)
```
Weitere gleichartige `not enough arguments`-Fehler folgen in `update_test.go` (Zeilen 62, 81, 93, 107, 119, 131, 148).

**Betroffene Dateien:** `internal/httpapi/update_test.go` und `internal/httpapi/flags_test.go` — beide deklarieren im selben Testpaket eine Funktion namens `doRequest` mit unterschiedlicher Signatur (`func(*http.ServeMux, string, string, string)` vs. `func(*testing.T, http.Handler, string, string, string)`). Eine der beiden Hilfsfunktionen muss umbenannt oder auf eine gemeinsame Signatur vereinheitlicht werden.

**Severity:** high