package main

import(
	"fmt"
	"log"
	"os/exec"
	"os"
	"strings"
	"gopkg.in/yaml.v3"
	"flag"
	"time"
	"bytes"
)

const (
    Reset  = "\033[0m"
    Green  = "\033[32m"
    Red    = "\033[31m"
    Yellow = "\033[33m"
)

func getActiveNodes(node_list *[]string) bool{
	*node_list = (*node_list)[:0]
	cmd := exec.Command("ros2", "node", "list")
	output, err := cmd.Output()

	if err != nil{
		log.Println("Error while executing ros2 command.", err)
		return false;
	}

	*node_list = strings.Split(string(output), "\n")
	if len(*node_list) > 1 && (*node_list)[len(*node_list) - 1] == ""{
		*node_list = (*node_list)[:len(*node_list) - 1]
	}
	return true
}

func getFilteringCommand(monitored_nodes *[]string) string{
	command_string := "ros2 node list"
	if len(*monitored_nodes) > 0{
		command_string = command_string + " | grep -E '"
		for _, monitored_node := range *monitored_nodes{
			command_string = command_string + monitored_node + "|"
		}
		command_string = command_string[:len(command_string)-1]
		command_string = command_string + "'"
		return command_string
	} else {
		return command_string
	}
}

func getActiveNodesFiltered(node_list *[]string, filtering_command *string) bool{
	*node_list = (*node_list)[:0]
	cmd := exec.Command("sh", "-c", *filtering_command)
	output, err := cmd.Output()

	if err != nil{
		log.Println("Error while executing ros2 command.", err)
		return false;
	}

	*node_list = strings.Split(string(output), "\n")
	if len(*node_list) > 1 && (*node_list)[len(*node_list) - 1] == ""{
		*node_list = (*node_list)[:len(*node_list) - 1]
	}
	return true
}

func clearScreen() {
    cmd := exec.Command("clear")
    cmd.Stdout = os.Stdout
    cmd.Run()
}

type Config struct {
    Monitored_nodes []string `yaml:"monitored_nodes"`
}

func main(){
	// Check for config path
	config_path_ptr := flag.String("config", "config.yaml", "path to the configuration file")

	flag.Parse()

	// Read and unmarshal the yaml config file
	var config Config
	yamlFile, err := os.ReadFile(*config_path_ptr)
 	if err != nil {
 		panic(err)
    }

	err = yaml.Unmarshal(yamlFile, &config)
 	if err != nil {
 		panic(err)
    }

	// Get slice of active nodes
	var node_list []string
	var output_buffer bytes.Buffer
	filtering_command := getFilteringCommand(&config.Monitored_nodes)
	
	/* 
		- Checks for active nodes that have similar names with the ones we want to monitor.
		- Compare the output with the nodes we want to monitor.
		- Buffer the results, clear the screen and output the buffer.
		- Sleep for X milliseconds
	*/

	for {
		//Check whether the requested nodes are present and print the output
		start := time.Now() //TODO remove
		if getActiveNodesFiltered(&node_list, &filtering_command) == false{
			fmt.Println("getting active nodes list failed!")
		}
		fmt.Println("--------------------------------------")
		for _, monitored_node := range config.Monitored_nodes {
			found := false
			for _, node := range node_list{
				if monitored_node == node{
					found = true
					break
				}
			}
			if found == true{
				output_buffer.WriteString(fmt.Sprintf("%s%s is running%s\n", Green, monitored_node, Reset))
			}else {
				output_buffer.WriteString(fmt.Sprintf("%s%s is not running%s\n", Red, monitored_node, Reset))
			}
		}
		duration := time.Since(start)// TODO remove
		clearScreen()
		output_buffer.WriteString(fmt.Sprintf("Elapsed time %v\n", duration)) //TODO remove
		fmt.Print(output_buffer.String())
		output_buffer.Reset()
		time.Sleep(1000 * time.Millisecond)
	}
}

/*TODOs
	- use signal handling for gracefully exiting
	- use go routines for more efficient processing
	- add other useful insights from a ros2 perspective, maybe topic Hz monitoring?
*/
