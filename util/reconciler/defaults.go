/*
Copyright 2020 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package reconciler implements the reconciler logic.
package reconciler

import (
	"encoding/binary"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"hash"
	"hash/fnv"
	"k8s.io/apimachinery/pkg/util/rand"
	"time"
)

const (
	// DefaultLoopTimeout is the default timeout for a reconcile loop (defaulted to the max ARM template duration).
	DefaultLoopTimeout = 90 * time.Minute
	// DefaultMappingTimeout is the default timeout for a controller request mapping func.
	DefaultMappingTimeout = 60 * time.Second
)

// DefaultedLoopTimeout will default the timeout if it is zero valued.
func DefaultedLoopTimeout(timeout time.Duration) time.Duration {
	if timeout <= 0 {
		return DefaultLoopTimeout
	}

	return timeout
}

func DeepHashObject(hasher hash.Hash, objectToWrite interface{}) {
	hasher.Reset()
	printer := spew.ConfigState{
		Indent:         " ",
		SortKeys:       true,
		DisableMethods: true,
		SpewKeys:       true,
	}

	//if _, err := printer.Fprintf(hasher, "%#v", objectToWrite); err != nil {
	//	return fmt.Errorf("failed to write object to hasher")
	//}
	//return nil

	printer.Fprintf(hasher, "%#v", objectToWrite)
}

func ComputeHash(obj interface{}, collisionCount *int32) string {
	objHasher := fnv.New32a()
	DeepHashObject(objHasher, obj)

	// Add collisionCount in the hash if it exists.
	if collisionCount != nil {
		collisionCountBytes := make([]byte, 8)
		binary.LittleEndian.PutUint32(collisionCountBytes, uint32(*collisionCount))
		objHasher.Write(collisionCountBytes)
	}

	return rand.SafeEncodeString(fmt.Sprint(objHasher.Sum32()))
}
