package main

import (
	"fmt"
	"net/http"
)

var html = `
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <title>Kubernetes Pod</title>
</head>
<body>
  <h3>Pod Info</h3>
  <ul>
    <li>Hostname: %s</li>
  </ul>
  <h3>Secret Details</h3>
  <ul>
    <li>Username: %s</li>
    <li>Password: %s</li>
  </ul>
</body>
</html>
`

func httpHandler(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, html, hostname, sl.username, sl.password)
}
