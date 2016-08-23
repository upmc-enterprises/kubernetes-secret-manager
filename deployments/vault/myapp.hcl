path "secret/*" {
  policy = "read"
}

path "auth/token/create*" {
  policy = "write"
}

path "auth/token/create-orphan*" {
  policy = "write"
}

path "sys/renew/mysql/creds/*/*" {
  policy = "write"
}

path "mysql/creds/*" {
  policy = "read"
}

path "mysql/roles/*" {
  policy = "read"
}
