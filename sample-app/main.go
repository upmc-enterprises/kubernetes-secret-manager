/*
Copyright (c) 2016, UPMC Enterprises
All rights reserved.
Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are met:
    * Redistributions of source code must retain the above copyright
      notice, this list of conditions and the following disclaimer.
    * Redistributions in binary form must reproduce the above copyright
      notice, this list of conditions and the following disclaimer in the
      documentation and/or other materials provided with the distribution.
    * Neither the name UPMC Enterprises nor the
      names of its contributors may be used to endorse or promote products
      derived from this software without specific prior written permission.
THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
DISCLAIMED. IN NO EVENT SHALL UPMC ENTERPRISES BE LIABLE FOR ANY
DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
(INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND
ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
*/

package main

import (
	"flag"
	"log"
	"net/http"
	"os"
)

var (
	httpAddr     string
	usernamePath string
	passwordPath string
)

var (
	hostname string
	sl       *SecretLoader
)

func main() {
	flag.StringVar(&httpAddr, "http", ":80", "HTTP Listen address.")
	flag.StringVar(&usernamePath, "username-secret-file", "/secrets/username", "Username secret path")
	flag.StringVar(&passwordPath, "password-secret-file", "/secrets/password", "Password secret path")

	flag.Parse()

	log.Println("Initializing application...")

	var err error
	sl, err = NewSecretLoader(usernamePath, passwordPath)
	if err != nil {
		log.Fatal(err)
	}
	hostname, err = os.Hostname()
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", httpHandler)

	log.Fatal(http.ListenAndServe(httpAddr, nil))
}
