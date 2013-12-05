// Copyright 2013 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package ssh_test

import (
	"io/ioutil"
	"os"
	"strings"
	stdtesting "testing"

	gc "launchpad.net/gocheck"

	coretesting "launchpad.net/juju-core/testing"
	"launchpad.net/juju-core/testing/testbase"
	"launchpad.net/juju-core/utils/ssh"
)

func Test(t *stdtesting.T) {
	gc.TestingT(t)
}

var validKey = `ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDEX/dPu4PmtvgK3La9zioCEDrJ` +
	`yUr6xEIK7Pr+rLgydcqWTU/kt7w7gKjOw4vvzgHfjKl09CWyvgb+y5dCiTk` +
	`9MxI+erGNhs3pwaoS+EavAbawB7iEqYyTep3YaJK+4RJ4OX7ZlXMAIMrTL+` +
	`UVrK89t56hCkFYaAgo3VY+z6rb/b3bDBYtE1Y2tS7C3au73aDgeb9psIrSV` +
	`86ucKBTl5X62FnYiyGd++xCnLB6uLximM5OKXfLzJQNS/QyZyk12g3D8y69` +
	`Xw1GzCSKX1u1+MQboyf0HJcG2ryUCLHdcDVppApyHx2OLq53hlkQ/yxdflD` +
	`qCqAE4j+doagSsIfC1T2T`

type AuthorisedKeysKeysSuite struct {
	testbase.LoggingSuite
}

var _ = gc.Suite(&AuthorisedKeysKeysSuite{})

func (s *AuthorisedKeysKeysSuite) SetUpTest(c *gc.C) {
	s.LoggingSuite.SetUpTest(c)
	fakeHome := coretesting.MakeEmptyFakeHomeWithoutJuju(c)
	s.AddCleanup(func(*gc.C) { fakeHome.Restore() })
}

func writeAuthKeysFile(c *gc.C, keys []string) {
	err := os.MkdirAll(coretesting.HomePath(".ssh"), 0755)
	c.Assert(err, gc.IsNil)
	authKeysFile := coretesting.HomePath(".ssh", "authorized_keys")
	err = ioutil.WriteFile(authKeysFile, []byte(strings.Join(keys, "\n")), 0644)
	c.Assert(err, gc.IsNil)
}

func (s *AuthorisedKeysKeysSuite) TestListKeys(c *gc.C) {
	keys := []string{
		validKey + " user@host",
		validKey + " anotheruser@host",
	}
	writeAuthKeysFile(c, keys)
	keys, err := ssh.ListKeys(ssh.KeyComments)
	c.Assert(err, gc.IsNil)
	c.Assert(keys, gc.DeepEquals, []string{"user@host", "anotheruser@host"})
}

func (s *AuthorisedKeysKeysSuite) TestListKeysFull(c *gc.C) {
	keys := []string{
		validKey + " user@host",
		validKey + " anotheruser@host",
	}
	writeAuthKeysFile(c, keys)
	actual, err := ssh.ListKeys(ssh.FullKeys)
	c.Assert(err, gc.IsNil)
	c.Assert(actual, gc.DeepEquals, keys)
}

func (s *AuthorisedKeysKeysSuite) TestAddNewKey(c *gc.C) {
	key := validKey + " user@host"
	err := ssh.AddKeys(key)
	c.Assert(err, gc.IsNil)
	actual, err := ssh.ListKeys(ssh.FullKeys)
	c.Assert(err, gc.IsNil)
	c.Assert(actual, gc.DeepEquals, []string{key})
}

func (s *AuthorisedKeysKeysSuite) TestAddMoreKeys(c *gc.C) {
	firstKey := validKey + " user@host"
	writeAuthKeysFile(c, []string{firstKey})
	moreKeys := []string{
		validKey + " anotheruser@host",
		validKey + " yetanotheruser@host",
	}
	err := ssh.AddKeys(moreKeys...)
	c.Assert(err, gc.IsNil)
	actual, err := ssh.ListKeys(ssh.FullKeys)
	c.Assert(err, gc.IsNil)
	c.Assert(actual, gc.DeepEquals, append([]string{firstKey}, moreKeys...))
}

func (s *AuthorisedKeysKeysSuite) TestAddDuplicateKey(c *gc.C) {
	key := validKey + " user@host"
	err := ssh.AddKeys(key)
	c.Assert(err, gc.IsNil)
	moreKeys := []string{
		validKey + " user@host",
		validKey + " yetanotheruser@host",
	}
	err = ssh.AddKeys(moreKeys...)
	c.Assert(err, gc.ErrorMatches, "cannot add duplicate ssh key: user@host")
}

func (s *AuthorisedKeysKeysSuite) TestAddKeyWithoutComment(c *gc.C) {
	keys := []string{
		validKey + " user@host",
		validKey,
	}
	err := ssh.AddKeys(keys...)
	c.Assert(err, gc.ErrorMatches, "cannot add ssh key without comment")
}

func (s *AuthorisedKeysKeysSuite) TestAddKeepsUnrecognised(c *gc.C) {
	writeAuthKeysFile(c, []string{validKey, "invalid-key"})
	anotherKey := validKey + " anotheruser@host"
	err := ssh.AddKeys(anotherKey)
	c.Assert(err, gc.IsNil)
	actual, err := ssh.ReadAuthorisedKeys()
	c.Assert(err, gc.IsNil)
	c.Assert(actual, gc.DeepEquals, []string{validKey, "invalid-key", anotherKey})
}

func (s *AuthorisedKeysKeysSuite) TestDeleteKeys(c *gc.C) {
	firstKey := validKey + " user@host"
	anotherKey := validKey + " anotheruser@host"
	writeAuthKeysFile(c, []string{firstKey, anotherKey})
	err := ssh.DeleteKeys("user@host")
	c.Assert(err, gc.IsNil)
	actual, err := ssh.ListKeys(ssh.FullKeys)
	c.Assert(err, gc.IsNil)
	c.Assert(actual, gc.DeepEquals, []string{anotherKey})
}

func (s *AuthorisedKeysKeysSuite) TestDeleteKeysKeepsUnregognised(c *gc.C) {
	firstKey := validKey + " user@host"
	writeAuthKeysFile(c, []string{firstKey, validKey, "invalid-key"})
	err := ssh.DeleteKeys("user@host")
	c.Assert(err, gc.IsNil)
	actual, err := ssh.ReadAuthorisedKeys()
	c.Assert(err, gc.IsNil)
	c.Assert(actual, gc.DeepEquals, []string{validKey, "invalid-key"})
}

func (s *AuthorisedKeysKeysSuite) TestDeleteNonExistentKey(c *gc.C) {
	firstKey := validKey + " user@host"
	writeAuthKeysFile(c, []string{firstKey})
	err := ssh.DeleteKeys("someone@host")
	c.Assert(err, gc.ErrorMatches, "cannot delete non existent key: someone@host")
}

func (s *AuthorisedKeysKeysSuite) TestDeleteLastKeyForbidden(c *gc.C) {
	keys := []string{
		validKey + " user@host",
		validKey + " yetanotheruser@host",
	}
	writeAuthKeysFile(c, keys)
	err := ssh.DeleteKeys("user@host", "yetanotheruser@host")
	c.Assert(err, gc.ErrorMatches, "cannot delete all keys")
}