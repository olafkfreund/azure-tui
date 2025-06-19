package main

import (
	"fmt"

	"github.com/olafkfreund/azure-tui/internal/azure/network"
)

func main() {
	fmt.Println("Testing Azure TUI Network Functions...")
	fmt.Println("=====================================")

	// Test Network Dashboard
	fmt.Println("\n1. Testing Network Dashboard:")
	fmt.Println("-----------------------------")
	dashboard := network.RenderNetworkDashboard()
	fmt.Println(dashboard)

	// Test Network Topology
	fmt.Println("\n2. Testing Network Topology:")
	fmt.Println("-----------------------------")
	topology := network.RenderNetworkTopology()
	fmt.Println(topology)

	// Test AI Network Analysis
	fmt.Println("\n3. Testing AI Network Analysis:")
	fmt.Println("-------------------------------")
	aiAnalysis := network.RenderNetworkAIAnalysis()
	fmt.Println(aiAnalysis)

	// Test VNet Details (with sample data)
	fmt.Println("\n4. Testing VNet Details:")
	fmt.Println("------------------------")
	vnetDetails := network.RenderVNetDetails("test-vnet", "test-rg")
	fmt.Println(vnetDetails)

	// Test NSG Details (with sample data)
	fmt.Println("\n5. Testing NSG Details:")
	fmt.Println("-----------------------")
	nsgDetails := network.RenderNSGDetails("test-nsg", "test-rg")
	fmt.Println(nsgDetails)

	fmt.Println("\n✅ All network functions executed successfully!")
}
