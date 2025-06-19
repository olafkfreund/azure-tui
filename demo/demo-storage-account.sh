#!/bin/bash

# Azure TUI Storage Account Demo Script
# Demonstrates the new Storage Account management capabilities

echo "💾 Azure TUI - Storage Account Management Demo"
echo "=============================================="
echo

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if Azure CLI is installed and logged in
echo -e "${BLUE}📋 Step 1: Checking Azure CLI setup...${NC}"
if ! command -v az &> /dev/null; then
    echo -e "${RED}❌ Azure CLI is not installed. Please install it first.${NC}"
    exit 1
fi

# Check if logged in
if ! az account show &> /dev/null; then
    echo -e "${YELLOW}⚠️  Not logged in to Azure. Please run 'az login' first.${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Azure CLI is installed and logged in${NC}"
echo

# List available storage accounts
echo -e "${BLUE}🗄️  Step 2: Available Storage Accounts...${NC}"
STORAGE_ACCOUNTS=$(az storage account list --query "[].{Name:name,ResourceGroup:resourceGroup,Location:location,Kind:kind}" --output table)
if [ -z "$STORAGE_ACCOUNTS" ]; then
    echo -e "${YELLOW}⚠️  No storage accounts found. Creating a demo storage account...${NC}"
    
    # Create a demo resource group and storage account
    DEMO_RG="azure-tui-demo-rg"
    DEMO_STORAGE="azuretuidemost$(date +%s)"
    
    echo "Creating resource group: $DEMO_RG"
    az group create --name "$DEMO_RG" --location "eastus" --output none
    
    echo "Creating storage account: $DEMO_STORAGE"
    az storage account create \
        --name "$DEMO_STORAGE" \
        --resource-group "$DEMO_RG" \
        --location "eastus" \
        --sku "Standard_LRS" \
        --kind "StorageV2" \
        --output none
    
    echo -e "${GREEN}✅ Demo storage account created: $DEMO_STORAGE${NC}"
    STORAGE_ACCOUNT="$DEMO_STORAGE"
else
    echo "$STORAGE_ACCOUNTS"
    # Use the first storage account found
    STORAGE_ACCOUNT=$(az storage account list --query "[0].name" --output tsv)
fi
echo

# Show detailed storage account information
echo -e "${BLUE}🔍 Step 3: Storage Account Details...${NC}"
echo "Selected Storage Account: $STORAGE_ACCOUNT"
az storage account show --name "$STORAGE_ACCOUNT" --query '{name:name,resourceGroup:resourceGroup,location:location,kind:kind,sku:sku.name,provisioningState:provisioningState,primaryLocation:primaryLocation,statusOfPrimary:statusOfPrimary}' --output table
echo

# List containers in the storage account
echo -e "${BLUE}📂 Step 4: Storage Containers...${NC}"
echo "Containers in '$STORAGE_ACCOUNT':"

# Get storage account key for container operations
STORAGE_KEY=$(az storage account keys list --account-name "$STORAGE_ACCOUNT" --query "[0].value" --output tsv)

# List containers
CONTAINERS=$(az storage container list --account-name "$STORAGE_ACCOUNT" --account-key "$STORAGE_KEY" --query "[].{Name:name,LastModified:properties.lastModified,PublicAccess:properties.publicAccess}" --output table 2>/dev/null)

if [ -z "$CONTAINERS" ] || [ "$CONTAINERS" = "Name    LastModified    PublicAccess" ]; then
    echo -e "${YELLOW}📭 No containers found. Creating demo containers...${NC}"
    
    # Create demo containers
    echo "Creating 'web-assets' container..."
    az storage container create --name "web-assets" --account-name "$STORAGE_ACCOUNT" --account-key "$STORAGE_KEY" --output none
    
    echo "Creating 'backup-data' container..."
    az storage container create --name "backup-data" --account-name "$STORAGE_ACCOUNT" --account-key "$STORAGE_KEY" --output none
    
    echo "Creating 'logs' container..."
    az storage container create --name "logs" --account-name "$STORAGE_ACCOUNT" --account-key "$STORAGE_KEY" --output none
    
    echo -e "${GREEN}✅ Demo containers created${NC}"
    
    # List containers again
    az storage container list --account-name "$STORAGE_ACCOUNT" --account-key "$STORAGE_KEY" --query "[].{Name:name,LastModified:properties.lastModified,PublicAccess:properties.publicAccess}" --output table
else
    echo "$CONTAINERS"
fi
echo

# Create demo blobs
echo -e "${BLUE}📄 Step 5: Demo Blobs...${NC}"

# Create temporary demo files
echo "<!DOCTYPE html><html><head><title>Demo</title></head><body><h1>Azure TUI Demo</h1></body></html>" > /tmp/index.html
echo "body { font-family: Arial, sans-serif; }" > /tmp/styles.css
echo "This is a demo text file for Azure TUI storage testing." > /tmp/readme.txt

echo "Uploading demo blobs to 'web-assets' container..."
az storage blob upload --account-name "$STORAGE_ACCOUNT" --account-key "$STORAGE_KEY" --container-name "web-assets" --name "index.html" --file "/tmp/index.html" --output none
az storage blob upload --account-name "$STORAGE_ACCOUNT" --account-key "$STORAGE_KEY" --container-name "web-assets" --name "styles.css" --file "/tmp/styles.css" --output none
az storage blob upload --account-name "$STORAGE_ACCOUNT" --account-key "$STORAGE_KEY" --container-name "web-assets" --name "readme.txt" --file "/tmp/readme.txt" --output none

echo "Blobs in 'web-assets' container:"
az storage blob list --account-name "$STORAGE_ACCOUNT" --account-key "$STORAGE_KEY" --container-name "web-assets" --query "[].{Name:name,Size:properties.contentLength,Type:properties.blobType,LastModified:properties.lastModified}" --output table

# Clean up temporary files
rm -f /tmp/index.html /tmp/styles.css /tmp/readme.txt
echo

# Show blob details
echo -e "${BLUE}📋 Step 6: Blob Details Example...${NC}"
echo "Detailed properties for 'index.html':"
az storage blob show --account-name "$STORAGE_ACCOUNT" --account-key "$STORAGE_KEY" --container-name "web-assets" --name "index.html" --query '{name:name,size:properties.contentLength,contentType:properties.contentType,lastModified:properties.lastModified,etag:properties.etag,blobType:properties.blobType}' --output table
echo

# Show Azure TUI usage instructions
echo -e "${GREEN}🎯 Step 7: Azure TUI Storage Management Usage${NC}"
echo "========================================"
echo
echo -e "${YELLOW}Launch Azure TUI and test Storage Account features:${NC}"
echo "1. Run: ./azure-tui"
echo "2. Navigate to a resource group containing storage accounts"
echo "3. Select storage account: '$STORAGE_ACCOUNT'"
echo
echo -e "${YELLOW}Storage Management Keyboard Shortcuts:${NC}"
echo "• [T]       - List Storage Containers"
echo "• [Shift+T] - Create Container"
echo "• [B]       - List Blobs in Container"
echo "• [U]       - Upload Blob"
echo "• [Ctrl+X]  - Delete Storage Item"
echo "• [d]       - Dashboard view"
echo "• [R]       - Refresh"
echo
echo -e "${YELLOW}Navigation Flow:${NC}"
echo "Storage Account → [T] → Container List → [B] → Blob List → [Enter] → Blob Details"
echo
echo -e "${GREEN}✅ Demo environment ready! Launch Azure TUI to test storage functionality.${NC}"
echo
echo -e "${BLUE}🧹 To clean up demo resources later:${NC}"
if [ ! -z "$DEMO_RG" ]; then
    echo "az group delete --name '$DEMO_RG' --yes --no-wait"
else
    echo "Demo containers were created in existing storage account '$STORAGE_ACCOUNT'"
    echo "You can delete them manually if needed:"
    echo "az storage container delete --name 'web-assets' --account-name '$STORAGE_ACCOUNT'"
    echo "az storage container delete --name 'backup-data' --account-name '$STORAGE_ACCOUNT'"
    echo "az storage container delete --name 'logs' --account-name '$STORAGE_ACCOUNT'"
fi
