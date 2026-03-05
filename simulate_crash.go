//go:build ignore

package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"time"
)

func runNode(role string, port string) *exec.Cmd {
	cmd := exec.Command("go", "run", "main.go", "-role="+role, "-port="+port)
	return cmd
}

func main() {
	var outs [5]*bytes.Buffer
	for i := 0; i < 5; i++ {
		outs[i] = &bytes.Buffer{}
	}

	cmd1 := runNode("worker", "8001")
	cmd1.Stdout = outs[0]
	cmd1.Stderr = outs[0]
	cmd1.Start()

	cmd2 := runNode("worker", "8002")
	cmd2.Stdout = outs[1]
	cmd2.Stderr = outs[1]
	cmd2.Start()

	cmd3 := runNode("worker", "8003")
	cmd3.Stdout = outs[2]
	cmd3.Stderr = outs[2]
	cmd3.Start()

	cmd4 := runNode("worker", "8004")
	cmd4.Stdout = outs[3]
	cmd4.Stderr = outs[3]
	cmd4.Start()

	// Start leader
	cmd5 := runNode("leader", "8000")
	cmd5.Stdout = outs[4]
	cmd5.Stderr = outs[4]
	cmd5.Start()

	fmt.Println("All nodes started. Waiting 5 seconds before simulated crash...")
	time.Sleep(5 * time.Second)

	fmt.Println("CRASHING LEADER (Node 5)...")
	cmd5.Process.Kill()
	cmd5.Wait()

	fmt.Println("Waiting 10 seconds for Bully algorithm to elect new leader...")
	time.Sleep(10 * time.Second)

	fmt.Println("CLEANING UP...")
	cmd1.Process.Kill()
	cmd2.Process.Kill()
	cmd3.Process.Kill()
	cmd4.Process.Kill()

	fmt.Println("\n--- LOGS ---")
	for i, out := range outs {
		fmt.Printf("\n=== Node %d Output ===\n", i+1)
		fmt.Println(out.String())
	}

	os.Exit(0)
}
