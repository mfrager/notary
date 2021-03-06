// Copyright 2018 Axel Wagner
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bytes"
	"crypto/sha512"
	"flag"
	"io"
	"log"
	"os"
	"strings"
	"encoding/hex"
	"errors"
	"fmt"

	config "github.com/mfrager/notary/internal/config"
	"github.com/mfrager/notary/roughtime"
)

func main() {
	verify := flag.Bool("verify", false, "verify a given chain")
	serversJSON := flag.String("servers", "", "server-list to use")
	hashParam := flag.String("hash", "", "pre-computed sha512 hash in hex")
	flag.Parse()

	if flag.NArg() < 1 {
		log.Fatalf("usage: %s [-servers <servers.json>] [-verify] [-hash <sha512>] <file>", os.Args[0])
		return
	}

	servers, err := serverList(*serversJSON)
	if err != nil {
		log.Fatal(err)
	}

    nonce, err := hashExternal(*hashParam, flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}

	if *verify {
		c, err := roughtime.LoadChain(os.Stdin)
		if err != nil {
			log.Fatal(err)
		}
		info, err := roughtime.VerifyChain(c, servers)
		if err != nil {
			log.Fatal(err)
		}
		if len(c.Links) == 0 || bytes.Compare(c.Links[0].NonceOrBlind, nonce) != 0 {
			log.Fatal("chain nonce does not match file")
		}
		fmt.Printf(info)
		return
	}

	if err := roughtime.Chain(os.Stdout, servers, nonce); err != nil {
		log.Fatal(err)
	}
}

func hashExternal(hashParam string, name string) ([]byte, error) {
	if len(hashParam) > 0 {
		if len(hashParam) == 128 {
			nonce, err := hex.DecodeString(hashParam)
			return nonce, err
		} else {
			return nil, errors.New("invalid sha512 hash value in hex")
		}
	} else {
		nonce, err := hashFile(name)
		return nonce, err
	}
}

func hashFile(name string) ([]byte, error) {
	f, err := os.Open(flag.Arg(0))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	h := sha512.New()
	if _, err = io.Copy(h, f); err != nil {
		return nil, err
	}
	return h.Sum(nil), nil
}

func serverList(name string) (*config.ServersJSON, error) {
	r := io.Reader(strings.NewReader(defaultServers))
	if name != "" {
		f, err := os.Open(name)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		r = f
	}
	return roughtime.ReadServersJSON(r)
}

var defaultServers = `{
	"servers": [
		{
			"name": "Google",
			"publicKeyType": "ed25519",
			"publicKey": "etPaaIxcBMY1oUeGpwvPMCJMwlRVNxv51KK/tktoJTQ=",
			"addresses": [
				{
					"protocol": "udp",
					"address": "roughtime.sandbox.google.com:2002"
				}
			]
		},
		{
			"name": "Cloudflare",
			"publicKeyType": "ed25519",
			"publicKey": "gD63hSj3ScS+wuOeGrubXlq35N1c5Lby/S+T7MNTjxo=",
			"addresses": [
				{
					"protocol": "udp",
					"address": "roughtime.cloudflare.com:2002"
				}
			]
		}
	]
}`
