// Copyright 2019 ChainSafe Systems (ON) Corp.
// This file is part of gossamer.
//
// The gossamer library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The gossamer library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the gossamer library. If not, see <http://www.gnu.org/licenses/>.

package trie

import (
	"fmt"

	"github.com/ChainSafe/gossamer/lib/common"
)

// ChildStorageKeyPrefix is the prefix for all child storage keys
var ChildStorageKeyPrefix = []byte(":child_storage:default:")

// PutChild inserts a child trie into the main trie at key :child_storage:[keyToChild]
func (t *Trie) PutChild(keyToChild []byte, child *Trie) error {
	childHash, err := child.Hash()
	if err != nil {
		return err
	}

	key := append(ChildStorageKeyPrefix, keyToChild...)
	value := [32]byte(childHash)

	t.Put(key, value[:])
	t.childTries[childHash] = child
	return nil
}

// GetChild returns the child trie at key :child_storage:[keyToChild]
func (t *Trie) GetChild(keyToChild []byte) (*Trie, error) {
	key := append(ChildStorageKeyPrefix, keyToChild...)
	childHash := t.Get(key)
	if childHash == nil {
		return nil, fmt.Errorf("child trie does not exist at key %s%s", ChildStorageKeyPrefix, keyToChild)
	}

	hash := [32]byte{}
	copy(hash[:], childHash)
	return t.childTries[common.Hash(hash)], nil
}

// PutIntoChild puts a key-value pair into the child trie located in the main trie at key :child_storage:[keyToChild]
func (t *Trie) PutIntoChild(keyToChild, key, value []byte) error {
	child, err := t.GetChild(keyToChild)
	if err != nil {
		return err
	}

	origChildHash, err := child.Hash()
	if err != nil {
		return err
	}

	child.Put(key, value)
	childHash, err := child.Hash()
	if err != nil {
		return err
	}

	t.childTries[origChildHash] = nil
	t.childTries[childHash] = child

	return t.PutChild(keyToChild, child)
}

// GetFromChild retrieves a key-value pair from the child trie located in the main trie at key :child_storage:[keyToChild]
func (t *Trie) GetFromChild(keyToChild, key []byte) ([]byte, error) {
	child, err := t.GetChild(keyToChild)
	if err != nil {
		return nil, err
	}

	if child == nil {
		return nil, fmt.Errorf("child trie does not exist at key %s%s", ChildStorageKeyPrefix, keyToChild)
	}

	val := child.Get(key)
	return val, nil
}

// DeleteChild deletes the child storage trie
func (t *Trie) DeleteChild(keyToChild []byte) {
	key := append(ChildStorageKeyPrefix, keyToChild...)
	t.Delete(key)
}

// ClearFromChild removes the child storage entry
func (t *Trie) ClearFromChild(keyToChild, key []byte) error {
	child, err := t.GetChild(keyToChild)
	if err != nil {
		return err
	}
	if child == nil {
		return fmt.Errorf("child trie does not exist at key %s%s", ChildStorageKeyPrefix, keyToChild)
	}
	child.Delete(key)
	return nil
}
