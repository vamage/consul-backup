package main

import (
	"fmt"
    "sort"
    "os"
    "io/ioutil"
    "strings"
	"github.com/hashicorp/consul/api"
    "github.com/docopt/docopt-go"
    "encoding/base64"
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

type ByCreateIndex api.KVPairs

func (a ByCreateIndex) Len() int           { return len(a) }
func (a ByCreateIndex) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
//Sort the KVs by createIndex
func (a ByCreateIndex) Less(i, j int) bool { return a[i].CreateIndex < a[j].CreateIndex }

//type ACLEntry struct {
//    CreateIndex uint64
//    ModifyIndex uint64
//    ID          string
//    Name        string
//    Type        string
//    Rules       string
//}

func backupKv(ipaddress string, token string, outfile string) {

    config := api.DefaultConfig()
    config.Address = ipaddress
    config.Token = token

	client, _ := api.NewClient(config)
	kv := client.KV()

	pairs, _, err := kv.List("/", nil)
    if err != nil {
        panic(err)
    }

    sort.Sort(ByCreateIndex(pairs))

    outstring := ""
	for _, element := range pairs {
        encoded_value := base64.StdEncoding.EncodeToString(element.Value)
        outstring += fmt.Sprintf("%s:%s\n", element.Key, encoded_value)
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
    // sort.Sort(ByCreateIndex(tokens))

    outstring := ""
	for _, element := range tokens {
        // outstring += fmt.Sprintf("%s:%s:%s:%s\n", element.ID, element.Name, element.Type, element.Rules)
        outstring += fmt.Sprintf("====\nID:%s\nName:%s\nType:%s\nRules:%s\n", element.ID, element.Name, element.Type, element.Rules)
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
func restoreKv(ipaddress string, token string, infile string) {

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
            decoded_value, decode_err := base64.StdEncoding.DecodeString(kvp[1])
            if decode_err != nil {
                panic(decode_err)
            }

            p := &api.KVPair{Key: kvp[0], Value: decoded_value}
            _, err := kv.Put(p, nil)
            if err != nil {
                panic(err)
            }
        }
    }
}

func restoreAcls(ipaddress string, token string, infile string) {

    config := api.DefaultConfig()
    config.Address = ipaddress
    config.Token = token

    data, err := ioutil.ReadFile(infile)
    if err != nil {
        panic(err)
    }

	client, _ := api.NewClient(config)
	acl := client.ACL()

    for _, block := range strings.Split(string(data), "====") {

        block = strings.TrimSpace(block)

        // first line can be empty, should abort
        if block == "" {
            continue
        }

        acl_rule_object := make(map[string]string)
        acl_block_part := ""

        for _, element := range strings.Split(string(block), "\n") {
            aclp := strings.SplitN(element, ":", 2)
            if len(aclp) > 1 && acl_block_part != "Rules" {
                acl_block_part = aclp[0]
                acl_rule_object[acl_block_part] = aclp[1]
            } else {
                acl_rule_object[acl_block_part] = acl_rule_object[acl_block_part] + "\n" + element
            }

        }

        acl_rule_object["Rules"] = strings.TrimSpace(string(acl_rule_object["Rules"]))

        e := &api.ACLEntry{
            ID: acl_rule_object["ID"],
            Name: acl_rule_object["Name"],
            Type: acl_rule_object["Type"],
            Rules: acl_rule_object["Rules"],
        }

        _, err := acl.Update(e, nil)
        if err != nil {
            panic(err)
        }
    }
}

func main() {

    usage := `Consul KV and ACL Backup with KV Restore tool.

Usage:
  consul-backup [-i IP:PORT] [-t TOKEN] [--kv] [--kvfile KVBACKUPFILE] [--acl] [--aclfile ACLBACKUPFILE] [--restore]
  consul-backup -h | --help
  consul-backup --version

Options:
  -h --help                          Show this screen.
  --version                          Show version.
  -i, --address=IP:PORT              The HTTP endpoint of Consul [default: 127.0.0.1:8500].
  -t, --token=TOKEN                  An ACL Token with proper permissions in Consul [default: ].
  -k, --kv                           Backup or restore KV
  -a, --acl                          Backup or restore ACL
  -K, --kvfile=KVBACKUPFILE          KV Backup Filename [default: kv.bkp].
  -A, --aclfile=ACLBACKUPFILE        ACL Backup Filename [default: acl.bkp].
  -r, --restore                      Activate restore mode`

  arguments, _ := docopt.Parse(usage, nil, true, "consul-backup 1.1", false)
  fmt.Println(arguments)

    if arguments["--restore"] == true {
        fmt.Println("Restore mode:")
        if arguments["--acl"] == false && arguments["--kv"] == false {
            fmt.Println("Nothing to do. Please specify '--kv' or '--acl' parameter")
            os.Exit(1)
        }
        fmt.Printf("Warning! This will overwrite existing kv or acls. Press [enter] to continue; CTL-C to exit")
        fmt.Scanln()
        if arguments["--kv"] == true {
            fmt.Println("Restoring KV from file: ", arguments["--kvfile"].(string))
            restoreKv(arguments["--address"].(string), arguments["--token"].(string), arguments["--kvfile"].(string))
        }
        if arguments["--acl"] == true {
            fmt.Println("Restoring ACL Tokens from file: ", arguments["--aclfile"].(string))
            restoreAcls(arguments["--address"].(string), arguments["--token"].(string), arguments["--aclfile"].(string))
        }
    } else {
        fmt.Println("Backup mode:")
        if arguments["--kv"] == true {
            fmt.Println("KV store will be backed up to file: ", arguments["--kvfile"].(string))
            backupKv(arguments["--address"].(string), arguments["--token"].(string), arguments["--kvfile"].(string))
        }
        if arguments["--acl"] == true {
            fmt.Println("ACL Tokens will be backed up to file: ", arguments["--aclfile"].(string))
            backupAcls(arguments["--address"].(string), arguments["--token"].(string), arguments["--aclfile"].(string))
        }
        if arguments["--acl"] == false && arguments["--kv"] == false {
            fmt.Println("Nothing to do. Please specify '--kv' or '--acl' parameter")
            os.Exit(1)
        }
    }
}
