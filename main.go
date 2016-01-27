package main

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strings"

	"github.com/docopt/docopt-go"
	"github.com/hashicorp/consul/api"
)

//type KVPair struct {
//    Key         string
//    CreateIndex uint64
//    ModifyIndex uint64
//    LockIndex   uint64
//    Flags       uint64
//    Value       []byte
//    Session     string
//}

type byCreateIndex api.KVPairs

func (a byCreateIndex) Len() int      { return len(a) }
func (a byCreateIndex) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

// Sort the KVs by createIndex
func (a byCreateIndex) Less(i, j int) bool { return a[i].CreateIndex < a[j].CreateIndex }

func backup(ipaddress string, token string, outfile string) {

	config := api.DefaultConfig()
	config.Address = ipaddress
	config.Token = token

	client, _ := api.NewClient(config)
	kv := client.KV()

	pairs, _, err := kv.List("/", nil)
	if err != nil {
		panic(err)
	}

	sort.Sort(byCreateIndex(pairs))

	outstring := ""
	for _, element := range pairs {
		encodedValue := base64.StdEncoding.EncodeToString(element.Value)
		outstring += fmt.Sprintf("%s:%s\n", element.Key, encodedValue)
	}

	file, err := os.Create(outfile)
	if err != nil {
		panic(err)
	}

	if _, err := file.Write([]byte(outstring)[:]); err != nil {
		panic(err)
	}
}

func backupAcls(ipaddress string, token string, outfile string) {

	config := api.DefaultConfig()
	config.Address = ipaddress
	config.Token = token

	client, _ := api.NewClient(config)
	acl := client.ACL()

	tokens, _, err := acl.List(nil)
	if err != nil {
		panic(err)
	}
	// sort.Sort(byCreateIndex(tokens))

	outstring := ""
	for _, element := range tokens {
		// outstring += fmt.Sprintf("%s:%s:%s:%s\n", element.ID, element.Name, element.Type, element.Rules)
		outstring += fmt.Sprintf("====\nID: %s\nName: %s\nType: %s\nRules:\n%s\n", element.ID, element.Name, element.Type, element.Rules)
	}

	file, err := os.Create(outfile)
	if err != nil {
		panic(err)
	}

	if _, err := file.Write([]byte(outstring)[:]); err != nil {
		panic(err)
	}
}

/* File needs to be in the following format:
   KEY1:VALUE1
   KEY2:VALUE2
*/
func restore(ipaddress string, token string, infile string) {

	config := api.DefaultConfig()
	config.Address = ipaddress
	config.Token = token

	data, err := ioutil.ReadFile(infile)
	if err != nil {
		panic(err)
	}

	client, _ := api.NewClient(config)
	kv := client.KV()

	for _, element := range strings.Split(string(data), "\n") {
		kvp := strings.Split(element, ":")

		if len(kvp) > 1 {
			decodedValue, decodeErr := base64.StdEncoding.DecodeString(kvp[1])
			if decodeErr != nil {
				panic(decodeErr)
			}

			p := &api.KVPair{Key: kvp[0], Value: decodedValue}
			_, err := kv.Put(p, nil)
			if err != nil {
				panic(err)
			}
		}
	}
}

func main() {

	usage := `Consul KV and ACL Backup with KV Restore tool.

Usage:
  consul-backup [-i IP:PORT] [-t TOKEN] [--aclbackup] [--aclbackupfile ACLBACKUPFILE] [--restore] <filename>
  consul-backup -h | --help
  consul-backup --version

Options:
  -h --help                          Show this screen.
  --version                          Show version.
  -i, --address=IP:PORT              The HTTP endpoint of Consul [default: 127.0.0.1:8500].
  -t, --token=TOKEN                  An ACL Token with proper permissions in Consul [default: ].
  -a, --aclbackup                    Backup ACLs, does nothing in restore mode. ACL restore not available at this time.
  -b, --aclbackupfile=ACLBACKUPFILE  ACL Backup Filename [default: acl.bkp].
  -r, --restore                      Activate restore mode`

	arguments, _ := docopt.Parse(usage, nil, true, "consul-backup 1.0", false)
	fmt.Println(arguments)

	if arguments["--restore"] == true {
		fmt.Println("Restore mode:")
		fmt.Printf("Warning! This will overwrite existing kv. Press [enter] to continue; CTL-C to exit")
		fmt.Scanln()
		fmt.Println("Restoring KV from file: ", arguments["<filename>"].(string))
		restore(arguments["--address"].(string), arguments["--token"].(string), arguments["<filename>"].(string))
	} else {
		fmt.Println("Backup mode:")
		fmt.Println("KV store will be backed up to file: ", arguments["<filename>"].(string))
		backup(arguments["--address"].(string), arguments["--token"].(string), arguments["<filename>"].(string))
		if arguments["--aclbackup"] == true {
			fmt.Println("ACL Tokens will be backed up to file: ", arguments["--aclbackupfile"].(string))
			backupAcls(arguments["--address"].(string), arguments["--token"].(string), arguments["--aclbackupfile"].(string))
		}
	}
}
