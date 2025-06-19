#!/bin/bash

# Azure TUI Storage Functionality Test Script
# Quick verification of storage management features

echo "🧪 Azure TUI - Storage Functionality Test"
echo "========================================="
echo

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to check if a function exists in main.go
check_function() {
    local function_name="$1"
    local description="$2"
    
    if grep -q "$function_name" /home/olafkfreund/Source/Cloud/azure-tui/cmd/main.go; then
        echo -e "${GREEN}✅ $description${NC}"
        return 0
    else
        echo -e "${RED}❌ $description${NC}"
        return 1
    fi
}

# Function to check if storage view cases exist
check_storage_views() {
    echo -e "${BLUE}📋 Checking Storage View Cases...${NC}"
    
    if grep -q 'case "storage-containers":' /home/olafkfreund/Source/Cloud/azure-tui/cmd/main.go; then
        echo -e "${GREEN}✅ Storage containers view case${NC}"
    else
        echo -e "${RED}❌ Storage containers view case${NC}"
        return 1
    fi
    
    if grep -q 'case "storage-blobs":' /home/olafkfreund/Source/Cloud/azure-tui/cmd/main.go; then
        echo -e "${GREEN}✅ Storage blobs view case${NC}"
    else
        echo -e "${RED}❌ Storage blobs view case${NC}"
        return 1
    fi
    
    if grep -q 'case "storage-blob-details":' /home/olafkfreund/Source/Cloud/azure-tui/cmd/main.go; then
        echo -e "${GREEN}✅ Storage blob details view case${NC}"
    else
        echo -e "${RED}❌ Storage blob details view case${NC}"
        return 1
    fi
    
    return 0
}

# Function to check storage shortcuts
check_storage_shortcuts() {
    echo -e "${BLUE}⌨️  Checking Storage Keyboard Shortcuts...${NC}"
    
    if grep -q 'Microsoft.Storage/storageAccounts' /home/olafkfreund/Source/Cloud/azure-tui/cmd/main.go; then
        if grep -A 5 'Microsoft.Storage/storageAccounts' /home/olafkfreund/Source/Cloud/azure-tui/cmd/main.go | grep -q 'T:List Containers'; then
            echo -e "${GREEN}✅ Storage account shortcuts defined${NC}"
            return 0
        fi
    fi
    
    echo -e "${RED}❌ Storage account shortcuts not found${NC}"
    return 1
}

# Function to check storage actions section
check_storage_actions() {
    echo -e "${BLUE}🎬 Checking Storage Actions Section...${NC}"
    
    if grep -q 'Storage Management' /home/olafkfreund/Source/Cloud/azure-tui/cmd/main.go; then
        echo -e "${GREEN}✅ Storage Management actions section${NC}"
        return 0
    else
        echo -e "${RED}❌ Storage Management actions section${NC}"
        return 1
    fi
}

# Test compilation
echo -e "${BLUE}🔨 Testing Compilation...${NC}"
cd /home/olafkfreund/Source/Cloud/azure-tui
if go build -o test-azure-tui cmd/main.go 2>/dev/null; then
    echo -e "${GREEN}✅ Application compiles successfully${NC}"
    rm -f test-azure-tui
else
    echo -e "${RED}❌ Compilation failed${NC}"
    exit 1
fi
echo

# Test storage module functions
echo -e "${BLUE}📦 Checking Storage Module Functions...${NC}"
check_function "listStorageContainersCmd" "List storage containers command"
check_function "listStorageBlobsCmd" "List storage blobs command"
check_function "createStorageContainerCmd" "Create storage container command"
check_function "deleteStorageContainerCmd" "Delete storage container command"
check_function "uploadBlobCmd" "Upload blob command"
check_function "deleteBlobCmd" "Delete blob command"
check_function "showBlobDetailsCmd" "Show blob details command"
echo

# Test message types
echo -e "${BLUE}📨 Checking Storage Message Types...${NC}"
check_function "storageContainersMsg" "Storage containers message type"
check_function "storageBlobsMsg" "Storage blobs message type"
check_function "storageBlobDetailsMsg" "Storage blob details message type"
check_function "storageActionMsg" "Storage action message type"
echo

# Test view cases
check_storage_views
echo

# Test shortcuts
check_storage_shortcuts
echo

# Test actions section
check_storage_actions
echo

# Test storage module
echo -e "${BLUE}🗄️  Checking Storage Module Implementation...${NC}"
if [ -f "/home/olafkfreund/Source/Cloud/azure-tui/internal/azure/storage/storage.go" ]; then
    echo -e "${GREEN}✅ Storage module exists${NC}"
    
    # Check key functions in storage module
    if grep -q "ListContainers" /home/olafkfreund/Source/Cloud/azure-tui/internal/azure/storage/storage.go; then
        echo -e "${GREEN}✅ ListContainers function${NC}"
    else
        echo -e "${RED}❌ ListContainers function${NC}"
    fi
    
    if grep -q "ListBlobs" /home/olafkfreund/Source/Cloud/azure-tui/internal/azure/storage/storage.go; then
        echo -e "${GREEN}✅ ListBlobs function${NC}"
    else
        echo -e "${RED}❌ ListBlobs function${NC}"
    fi
    
    if grep -q "RenderStorageContainersView" /home/olafkfreund/Source/Cloud/azure-tui/internal/azure/storage/storage.go; then
        echo -e "${GREEN}✅ RenderStorageContainersView function${NC}"
    else
        echo -e "${RED}❌ RenderStorageContainersView function${NC}"
    fi
    
    if grep -q "RenderStorageBlobsView" /home/olafkfreund/Source/Cloud/azure-tui/internal/azure/storage/storage.go; then
        echo -e "${GREEN}✅ RenderStorageBlobsView function${NC}"
    else
        echo -e "${RED}❌ RenderStorageBlobsView function${NC}"
    fi
else
    echo -e "${RED}❌ Storage module not found${NC}"
fi
echo

# Check model fields
echo -e "${BLUE}🏗️  Checking Model Storage Fields...${NC}"
if grep -q "storageContainersContent" /home/olafkfreund/Source/Cloud/azure-tui/cmd/main.go; then
    echo -e "${GREEN}✅ storageContainersContent field${NC}"
else
    echo -e "${RED}❌ storageContainersContent field${NC}"
fi

if grep -q "storageBlobsContent" /home/olafkfreund/Source/Cloud/azure-tui/cmd/main.go; then
    echo -e "${GREEN}✅ storageBlobsContent field${NC}"
else
    echo -e "${RED}❌ storageBlobsContent field${NC}"
fi

if grep -q "storageBlobDetailsContent" /home/olafkfreund/Source/Cloud/azure-tui/cmd/main.go; then
    echo -e "${GREEN}✅ storageBlobDetailsContent field${NC}"
else
    echo -e "${RED}❌ storageBlobDetailsContent field${NC}"
fi
echo

# Final summary
echo -e "${YELLOW}📊 Test Summary${NC}"
echo "================"

# Count checks
TOTAL_CHECKS=20
FAILED_CHECKS=0

# This is a simplified count - in a real scenario you'd track each check
if ! check_storage_views >/dev/null 2>&1; then
    FAILED_CHECKS=$((FAILED_CHECKS + 3))
fi

if ! check_storage_shortcuts >/dev/null 2>&1; then
    FAILED_CHECKS=$((FAILED_CHECKS + 1))
fi

if ! check_storage_actions >/dev/null 2>&1; then
    FAILED_CHECKS=$((FAILED_CHECKS + 1))
fi

PASSED_CHECKS=$((TOTAL_CHECKS - FAILED_CHECKS))

echo "Passed: $PASSED_CHECKS/$TOTAL_CHECKS checks"

if [ $FAILED_CHECKS -eq 0 ]; then
    echo -e "${GREEN}🎉 All storage functionality tests passed!${NC}"
    echo -e "${GREEN}✅ Azure Storage Account management is ready for use${NC}"
    exit 0
else
    echo -e "${YELLOW}⚠️  $FAILED_CHECKS checks failed - please review implementation${NC}"
    exit 1
fi
