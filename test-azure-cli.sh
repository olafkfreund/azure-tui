#!/bin/bash

echo "Testing Azure CLI integration..."

# Test 1: Check if Azure CLI is available
echo "1. Testing Azure CLI availability:"
if command -v az &> /dev/null; then
    echo "✓ Azure CLI is installed"
else
    echo "✗ Azure CLI is not installed"
    exit 1
fi

# Test 2: Check if user is logged in
echo -e "\n2. Testing Azure authentication:"
if az account show &> /dev/null; then
    echo "✓ User is logged into Azure"
    echo "Current subscription:"
    az account show --query "name" -o tsv
else
    echo "✗ User is not logged into Azure"
    echo "Please run: az login"
    exit 1
fi

# Test 3: List subscriptions
echo -e "\n3. Testing subscription listing:"
sub_count=$(az account list --query "length(@)" -o tsv)
echo "✓ Found $sub_count subscriptions"

# Test 4: List resource groups
echo -e "\n4. Testing resource group listing:"
rg_count=$(az group list --query "length(@)" -o tsv)
echo "✓ Found $rg_count resource groups"

echo -e "\n✓ All Azure CLI tests passed!"
echo "The application should be able to load Azure data."
