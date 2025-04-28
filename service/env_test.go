// The MIT License (MIT)

// Copyright (c) 2020, Digital Scarcity

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.
package service_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	eostest "github.com/digital-scarcity/eos-go-test"

	eos "github.com/sebastianmontero/eos-go"
	"gotest.tools/assert"
)

type Environment struct {
	X     context.Context
	A     *eos.API
	CRP   time.Duration
	Users []eos.AccountName
}

const testingEndpoint = "http://localhost:8888"

var E *Environment

func NewEnvironment() (*Environment, error) {

	api := eos.New(testingEndpoint)
	api.Debug = true

	ctx := context.Background()
	keyBag := &eos.KeyBag{}
	err := keyBag.ImportPrivateKey(ctx, eostest.DefaultKey())
	if err != nil {
		return nil, err
	}
	api.SetSigner(keyBag)

	return &Environment{
		A:   api,
		X:   ctx,
		CRP: time.Millisecond * 300,
	}, nil

}
func (m *Environment) Setup(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)
	m.setupEnvironment(t)
}

func setupTestCase(t *testing.T) func(t *testing.T) {
	t.Log("Bootstrapping testing environment ...")

	cmd, err := eostest.RestartNodeos(false,
		"-e", "-p", "eosio",
		"--plugin", "eosio::producer_plugin",
		"--plugin", "eosio::producer_api_plugin",
		"--plugin", "eosio::chain_api_plugin",
		"--plugin", "eosio::http_plugin",
		"--plugin", "eosio::trace_api_plugin",
		"--trace-no-abis",
		// "--plugin", "eosio::history_plugin",
		// "--plugin", "eosio::history_api_plugin",
		// `--filter-on="*"`,
		"--access-control-allow-origin", "*",
		"--contracts-console",
		"--http-validate-host", "false",
		"--verbose-http-errors",
		"--delete-all-blocks",
		// "--logconf", "/home/sebastian/nodeos/logging.json"
	)
	assert.NilError(t, err)

	// outfile, err := os.Create("/home/sebastian/nodeos/nodeos.log")
	// if err != nil {
	// 	panic(err)
	// }
	// defer outfile.Close()
	// // cmd.Stdout = outfile
	// cmd.Stderr = outfile

	t.Log("nodeos PID: ", cmd.Process.Pid)

	return func(t *testing.T) {
		// any cleanup code would go here
	}
}

func (m *Environment) setupEnvironment(t *testing.T) {
	numUsers := 7

	m.Users = make([]eos.AccountName, 0, numUsers)
	for i := 0; i < numUsers; i++ {

		user, err := eostest.CreateAccountFromString(m.X, m.A, fmt.Sprintf("user%c", rune('a'+i)), eostest.DefaultKey())
		assert.NilError(t, err)

		m.Users = append(m.Users, user)
	}
}

func (m *Environment) deployContract(t *testing.T, contract, wasm, abi string) eos.AccountName {
	account, err := eostest.CreateAccountFromString(m.X, m.A, contract, eostest.DefaultKey())
	assert.NilError(t, err)

	_, err = eostest.SetContract(m.X, m.A, account, wasm, abi)
	assert.NilError(t, err)
	return account
}

func TestMain(m *testing.M) {
	beforeAll()
	// exec test and this returns an exit code to pass to os
	retCode := m.Run()
	afterAll()
	// If exit code is distinct of zero,
	// the test will be failed (red)
	os.Exit(retCode)
}

func beforeAll() {
	var err error
	E, err = NewEnvironment()
	if err != nil {
		log.Fatal("Failed to create Envionment", err)
	}
}

func afterAll() {
}
