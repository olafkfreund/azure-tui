#!/bin/bash

# Terraform Integration Test Script
# This script tests the Terraform integration for azure-tui

set -e

echo "🚀 Testing Terraform Integration for Azure-TUI"
echo "=============================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test variables
TEST_DIR="/tmp/aztui-terraform-test"
TERRAFORM_DIR="terraform"
WORKSPACE_NAME="test-workspace"
TEMPLATE_NAME="linux-vm"

# Cleanup function
cleanup() {
    echo -e "${YELLOW}🧹 Cleaning up test environment...${NC}"
    if [ -d "$TEST_DIR" ]; then
        rm -rf "$TEST_DIR"
    fi
}

# Set up cleanup on exit
trap cleanup EXIT

# Function to print test status
print_status() {
    local status=$1
    local message=$2
    if [ "$status" = "PASS" ]; then
        echo -e "${GREEN}✅ PASS: $message${NC}"
    elif [ "$status" = "FAIL" ]; then
        echo -e "${RED}❌ FAIL: $message${NC}"
    elif [ "$status" = "INFO" ]; then
        echo -e "${BLUE}ℹ️  INFO: $message${NC}"
    elif [ "$status" = "WARN" ]; then
        echo -e "${YELLOW}⚠️  WARN: $message${NC}"
    fi
}

# Function to run test and capture result
run_test() {
    local test_name=$1
    local test_command=$2
    
    echo -e "\n${BLUE}🧪 Running test: $test_name${NC}"
    
    if eval "$test_command"; then
        print_status "PASS" "$test_name"
        return 0
    else
        print_status "FAIL" "$test_name"
        return 1
    fi
}

# Test 1: Check if Terraform is installed
test_terraform_installed() {
    if command -v terraform &> /dev/null; then
        local version=$(terraform --version | head -n1)
        print_status "INFO" "Terraform found: $version"
        return 0
    else
        print_status "WARN" "Terraform not installed - some tests will be skipped"
        return 1
    fi
}

# Test 2: Check Terraform template structure
test_template_structure() {
    local templates_found=0
    
    # Check if terraform directory exists
    if [ ! -d "$TERRAFORM_DIR" ]; then
        print_status "FAIL" "Terraform directory not found"
        return 1
    fi
    
    # Check template categories
    for category in "vm" "sql" "aks" "aci"; do
        if [ -d "$TERRAFORM_DIR/templates/$category" ]; then
            templates_found=$((templates_found + 1))
            print_status "PASS" "Found $category templates"
        else
            print_status "FAIL" "Missing $category templates"
        fi
    done
    
    if [ $templates_found -eq 4 ]; then
        return 0
    else
        return 1
    fi
}

# Test 3: Validate template files
test_template_files() {
    local template_dir="$TERRAFORM_DIR/templates/vm/linux-vm"
    
    if [ ! -d "$template_dir" ]; then
        print_status "FAIL" "Linux VM template directory not found"
        return 1
    fi
    
    # Check required files
    local required_files=("main.tf" "variables.tf" "outputs.tf")
    local files_found=0
    
    for file in "${required_files[@]}"; do
        if [ -f "$template_dir/$file" ]; then
            files_found=$((files_found + 1))
            print_status "PASS" "Found $file"
        else
            print_status "FAIL" "Missing $file"
        fi
    done
    
    if [ $files_found -eq 3 ]; then
        return 0
    else
        return 1
    fi
}

# Test 4: Validate Terraform syntax
test_terraform_syntax() {
    if ! command -v terraform &> /dev/null; then
        print_status "WARN" "Terraform not installed - skipping syntax validation"
        return 0
    fi
    
    local template_dir="$TERRAFORM_DIR/templates/vm/linux-vm"
    
    # Create temporary directory for testing
    mkdir -p "$TEST_DIR/syntax-test"
    cp -r "$template_dir"/* "$TEST_DIR/syntax-test/"
    
    cd "$TEST_DIR/syntax-test"
    
    # Initialize and validate
    if terraform init -backend=false > /dev/null 2>&1; then
        print_status "PASS" "Terraform init successful"
    else
        print_status "FAIL" "Terraform init failed"
        return 1
    fi
    
    if terraform validate > /dev/null 2>&1; then
        print_status "PASS" "Terraform syntax validation passed"
        return 0
    else
        print_status "FAIL" "Terraform syntax validation failed"
        return 1
    fi
}

# Test 5: Check Go module compilation
test_go_compilation() {
    echo -e "\n${BLUE}🔨 Testing Go module compilation...${NC}"
    
    # Try to build the main module
    if go build -o "$TEST_DIR/aztui-test" ./cmd/main.go > /dev/null 2>&1; then
        print_status "PASS" "Go module compilation successful"
        return 0
    else
        print_status "FAIL" "Go module compilation failed"
        echo "Run 'go build ./cmd/main.go' for detailed error information"
        return 1
    fi
}

# Test 6: Check Terraform integration imports
test_terraform_imports() {
    echo -e "\n${BLUE}📦 Testing Terraform package imports...${NC}"
    
    # Check if terraform package can be imported
    if go list ./internal/terraform > /dev/null 2>&1; then
        print_status "PASS" "Terraform package imports successfully"
    else
        print_status "FAIL" "Terraform package import failed"
        return 1
    fi
    
    # Check specific terraform files
    local terraform_files=("terraform.go" "tui.go" "commands.go" "integration.go")
    local files_found=0
    
    for file in "${terraform_files[@]}"; do
        if [ -f "internal/terraform/$file" ]; then
            files_found=$((files_found + 1))
            print_status "PASS" "Found internal/terraform/$file"
        else
            print_status "WARN" "Missing internal/terraform/$file"
        fi
    done
    
    return 0
}

# Test 7: Configuration validation
test_configuration() {
    echo -e "\n${BLUE}⚙️  Testing configuration system...${NC}"
    
    # Check if example config exists
    if [ -f "$TERRAFORM_DIR/config.yaml.example" ]; then
        print_status "PASS" "Example configuration file found"
    else
        print_status "WARN" "Example configuration file missing"
    fi
    
    # Test config structure by checking if it's valid YAML
    if [ -f "$TERRAFORM_DIR/config.yaml.example" ]; then
        if command -v yq &> /dev/null; then
            if yq eval '.' "$TERRAFORM_DIR/config.yaml.example" > /dev/null 2>&1; then
                print_status "PASS" "Configuration file has valid YAML syntax"
            else
                print_status "FAIL" "Configuration file has invalid YAML syntax"
                return 1
            fi
        else
            print_status "INFO" "yq not available - skipping YAML validation"
        fi
    fi
    
    return 0
}

# Test 8: Template variable validation
test_template_variables() {
    echo -e "\n${BLUE}📋 Testing template variables...${NC}"
    
    local template_dir="$TERRAFORM_DIR/templates/vm/linux-vm"
    local variables_file="$template_dir/variables.tf"
    
    if [ ! -f "$variables_file" ]; then
        print_status "FAIL" "Variables file not found"
        return 1
    fi
    
    # Check for required variables
    local required_vars=("resource_group_name" "location" "vm_name")
    local vars_found=0
    
    for var in "${required_vars[@]}"; do
        if grep -q "variable \"$var\"" "$variables_file"; then
            vars_found=$((vars_found + 1))
            print_status "PASS" "Found variable: $var"
        else
            print_status "FAIL" "Missing variable: $var"
        fi
    done
    
    if [ $vars_found -eq ${#required_vars[@]} ]; then
        return 0
    else
        return 1
    fi
}

# Test 9: AI Integration compatibility
test_ai_integration() {
    echo -e "\n${BLUE}🤖 Testing AI integration compatibility...${NC}"
    
    # Check if AI package exists
    if [ -d "internal/openai" ]; then
        print_status "PASS" "AI package found"
    else
        print_status "FAIL" "AI package not found"
        return 1
    fi
    
    # Check if config includes AI settings
    if go run -c 'package main; import "github.com/olafkfreund/azure-tui/internal/config"; func main() {}' 2>/dev/null; then
        print_status "PASS" "Config package accessible for AI integration"
    else
        print_status "WARN" "Config package may have issues"
    fi
    
    return 0
}

# Test 10: Integration test summary
test_integration_summary() {
    echo -e "\n${BLUE}📊 Integration Test Summary${NC}"
    echo "=============================="
    
    local total_templates=0
    local categories=("vm" "sql" "aks" "aci")
    
    for category in "${categories[@]}"; do
        if [ -d "$TERRAFORM_DIR/templates/$category" ]; then
            local count=$(find "$TERRAFORM_DIR/templates/$category" -name "main.tf" | wc -l)
            total_templates=$((total_templates + count))
            print_status "INFO" "$category category: $count templates"
        fi
    done
    
    print_status "INFO" "Total templates found: $total_templates"
    
    # Check terraform directory structure
    if [ -d "$TERRAFORM_DIR/templates" ] && [ -f "$TERRAFORM_DIR/README.md" ]; then
        print_status "PASS" "Terraform integration structure is complete"
        return 0
    else
        print_status "FAIL" "Terraform integration structure is incomplete"
        return 1
    fi
}

# Main test execution
main() {
    echo -e "${BLUE}Starting Terraform Integration Tests...${NC}\n"
    
    local test_results=()
    local test_count=0
    local passed_count=0
    
    # List of all tests
    tests=(
        "Terraform Installation:test_terraform_installed"
        "Template Structure:test_template_structure"
        "Template Files:test_template_files" 
        "Terraform Syntax:test_terraform_syntax"
        "Go Compilation:test_go_compilation"
        "Terraform Imports:test_terraform_imports"
        "Configuration:test_configuration"
        "Template Variables:test_template_variables"
        "AI Integration:test_ai_integration"
        "Integration Summary:test_integration_summary"
    )
    
    # Run each test
    for test in "${tests[@]}"; do
        test_name=$(echo "$test" | cut -d: -f1)
        test_function=$(echo "$test" | cut -d: -f2)
        
        test_count=$((test_count + 1))
        
        if run_test "$test_name" "$test_function"; then
            passed_count=$((passed_count + 1))
            test_results+=("PASS: $test_name")
        else
            test_results+=("FAIL: $test_name")
        fi
    done
    
    # Print final results
    echo -e "\n${BLUE}===============================================${NC}"
    echo -e "${BLUE}🏁 FINAL TEST RESULTS${NC}"
    echo -e "${BLUE}===============================================${NC}"
    
    for result in "${test_results[@]}"; do
        if [[ $result == PASS* ]]; then
            echo -e "${GREEN}✅ $result${NC}"
        else
            echo -e "${RED}❌ $result${NC}"
        fi
    done
    
    echo -e "\n${BLUE}Summary: $passed_count/$test_count tests passed${NC}"
    
    if [ $passed_count -eq $test_count ]; then
        echo -e "${GREEN}🎉 All tests passed! Terraform integration is ready.${NC}"
        exit 0
    elif [ $passed_count -gt $((test_count / 2)) ]; then
        echo -e "${YELLOW}⚠️  Most tests passed. Some minor issues to address.${NC}"
        exit 1
    else
        echo -e "${RED}❌ Multiple test failures. Integration needs work.${NC}"
        exit 2
    fi
}

# Run the main function
main "$@"
