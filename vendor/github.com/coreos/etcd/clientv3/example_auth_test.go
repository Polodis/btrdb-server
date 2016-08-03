// Copyright 2016 The etcd Authors
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

package clientv3_test

import (
	"fmt"
	"log"

	"github.com/coreos/etcd/clientv3"
	"golang.org/x/net/context"
)

func ExampleAuth() {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: dialTimeout,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer cli.Close()

	authapi := clientv3.NewAuth(cli)

	if _, err = authapi.RoleAdd(context.TODO(), "root"); err != nil {
		log.Fatal(err)
	}

	if _, err = authapi.RoleGrantPermission(
		context.TODO(),
		"root", // role name
		"foo",  // key
		"zoo",  // range end
		clientv3.PermissionType(clientv3.PermReadWrite),
	); err != nil {
		log.Fatal(err)
	}

	if _, err = authapi.UserAdd(context.TODO(), "root", "123"); err != nil {
		log.Fatal(err)
	}

	if _, err = authapi.UserGrantRole(context.TODO(), "root", "root"); err != nil {
		log.Fatal(err)
	}

	if _, err = authapi.AuthEnable(context.TODO()); err != nil {
		log.Fatal(err)
	}

	cliAuth, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: dialTimeout,
		Username:    "root",
		Password:    "123",
	})
	if err != nil {
		log.Fatal(err)
	}
	defer cliAuth.Close()

	kv := clientv3.NewKV(cliAuth)
	if _, err = kv.Put(context.TODO(), "foo1", "bar"); err != nil {
		log.Fatal(err)
	}

	_, err = kv.Txn(context.TODO()).
		If(clientv3.Compare(clientv3.Value("zoo1"), ">", "abc")).
		Then(clientv3.OpPut("zoo1", "XYZ")).
		Else(clientv3.OpPut("zoo1", "ABC")).
		Commit()
	fmt.Println(err)

	// now check the permission
	authapi2 := clientv3.NewAuth(cliAuth)
	resp, err := authapi2.RoleGet(context.TODO(), "root")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("root user permission: key %q, range end %q\n", resp.Perm[0].Key, resp.Perm[0].RangeEnd)

	if _, err = authapi2.AuthDisable(context.TODO()); err != nil {
		log.Fatal(err)
	}
	// Output: etcdserver: permission denied
	// root user permission: key "foo", range end "zoo"
}