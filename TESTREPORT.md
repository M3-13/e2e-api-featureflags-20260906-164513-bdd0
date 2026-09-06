VERDICT: BUGS_FOUND

- **Title:** Doppelt deklarierte Test-Hilfsfunktion `doRequest` bricht den Test-Build
- **Symptom:** `go test ./...` schlägt fehl. Das Paket `featureflags/internal/httpapi` kann nicht kompiliert werden, dadurch läuft die gesamte Test-Suite nicht grün (AC-10 nicht erfüllt).
- **Repro:** Im Projektverzeichnis `go test ./...` ausführen.
- **Evidence:**
  ```
  FAIL	featureflags/internal/httpapi [build failed]
  internal\httpapi\helpers_test.go:10:6: doRequest redeclared in this block
  	internal\httpapi\flags_test.go:49:6: other declaration of doRequest
  ```
- **Suspected file(s):** `internal/httpapi/helpers_test.go` und `internal/httpapi/flags_test.go` — beide Dateien deklarieren eine Funktion `doRequest` im selben Paket. Gemeinsame Ursache ist die doppelte Deklaration; eine der beiden Funktionen muss umbenannt oder entfernt werden.
- **Severity:** high