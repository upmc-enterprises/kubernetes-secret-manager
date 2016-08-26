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
	"crypto/tls"
	"io/ioutil"
	"log"
	"sync"

	"golang.org/x/exp/inotify"
)

type SecretLoader struct {
	sync.RWMutex
	username           string
	password           string
	usernameSecretPath string
	passwordSecretPath string
	Error              chan error
	watcher            *inotify.Watcher
}

func NewSecretLoader(usernamePath, passwordPath string) (*SecretLoader, error) {
	sl := &SecretLoader{
		usernameSecretPath: usernamePath,
		passwordSecretPath: passwordPath,
		Error:              make(chan error, 10),
	}
	err := sl.setSecret()
	if err != nil {
		return nil, err
	}

	go sl.watchSecrets()

	return sl, nil
}

func (sl *SecretLoader) GetSecrets(clientHello *tls.ClientHelloInfo) (string, string, error) {
	sl.RLock()
	defer sl.RUnlock()
	return sl.username, sl.password, nil
}

func (sl *SecretLoader) setSecret() error {
	log.Println("Loading secrets...")
	user, err := ioutil.ReadFile(sl.usernameSecretPath)
	if err != nil {
		return err
	}

	pass, err := ioutil.ReadFile(sl.passwordSecretPath)
	if err != nil {
		return err
	}

	sl.Lock()
	sl.username = string(user)
	sl.password = string(pass)
	sl.Unlock()
	return nil
}

func (sl *SecretLoader) watchSecrets() error {
	log.Println("Watching for secret changes...")
	err := sl.newWatcher()
	if err != nil {
		return err
	}

	for {
		select {
		case <-sl.watcher.Event:
			log.Println("Reloading secrets...")
			err := sl.setSecret()
			if err != nil {
				sl.Error <- err
			}
			log.Println("Reloading secrets complete.")
			err = sl.resetWatcher()
			if err != nil {
				sl.Error <- err
			}
		case err := <-sl.watcher.Error:
			sl.Error <- err
		}
	}
}

func (sl *SecretLoader) newWatcher() error {
	var err error
	sl.watcher, err = inotify.NewWatcher()
	if err != nil {
		return err
	}
	err = sl.watcher.AddWatch(sl.usernameSecretPath, inotify.IN_IGNORED)
	if err != nil {
		return err
	}
	return sl.watcher.AddWatch(sl.passwordSecretPath, inotify.IN_IGNORED)
}

func (sl *SecretLoader) resetWatcher() error {
	err := sl.watcher.Close()
	if err != nil {
		return err
	}
	return sl.newWatcher()
}
