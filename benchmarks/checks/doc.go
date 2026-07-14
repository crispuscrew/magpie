// Package checks holds one injected check per bench task: each file is tagged
// `taskcheck`, declares package notekeep, and is copied into the fixture work
// tree as taskcheck_test.go — invisible to normal builds of this repo.
package checks
