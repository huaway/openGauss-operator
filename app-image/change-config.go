package main

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

func main() {
	args := os.Args
	if len(args) != 4 {
		fmt.Println("Not enough parameter")
		return
	}

	content, err := ioutil.ReadFile("../example/opengauss-new.yaml")
	if err != nil {
		fmt.Println("Unable to open file shardingsphere-proxy/config-readwrite-splitting.yaml")
		return
	}	
	
	var ret string
	if args[1] == "0.25" {
		ret = strings.Replace(string(content), "cpu:", "cpu: 250m", 2)
	} else if args[1] == "0.5" {
		ret = strings.Replace(string(content), "cpu:", "cpu: 500m", 2)
	} else {
		ret = strings.Replace(string(content), "cpu:", fmt.Sprintf("cpu: %s000m", args[1]), 2)
	}

	if args[2] == "0.5" {
		ret = strings.Replace(string(ret), "memory:", "memory: 500Mi", 2)
	} else {
		ret = strings.Replace(string(ret), "memory:", fmt.Sprintf("memory: %sGi", args[2]), 2)
	}
	ret = strings.Replace(string(ret), "replicas:", fmt.Sprintf("replicas: %s", args[3]), 1)

	ioutil.WriteFile("../example/opengauss-test.yaml", []byte(ret), fs.FileMode(os.O_WRONLY))

	// modify buffer pool
	content, err = ioutil.ReadFile("../configs/config-base.yaml")
	if err != nil {
		fmt.Println("Unable to open file configs/config-small.yaml")
		return
	}	
	
	configs := make(map[int]string)
	configs[0] = "256MB"
	configs[1] = "512MB"
	configs[2] = "1GB"
	configs[3] = "1536MB"
	configs[4] = "2GB"
	configs[5] = "2560MB"
	configs[6] = "3GB"
	configs[7] = "3584MB"
	configs[8] = "4GB"
	configs[9] = "4608MB"
	configs[10] = "5GB"
	configs[11] = "5632MB"
	configs[12] = "6GB"
	configs[13] = "6656MB"
	configs[14] = "7GB"
	configs[15] = "7680MB"
	configs[16] = "8GB"
	configs[32] = "16GB"

	var memory int
	if args[2] == "0.5" {
		memory = 0
	} else {
		memory ,_ = strconv.Atoi(args[2])
	}
	
	
	ret = strings.Replace(string(content), "= hahaha", fmt.Sprintf("= %s", configs[memory]), -1)
	
	ioutil.WriteFile("../configs/config-small.yaml", []byte(ret), fs.FileMode(os.O_WRONLY))
}
