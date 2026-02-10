#!/bin/bash

# --- CONFIGURATION ---
BINARY_NAME="creep-analyzer"
MAIN_FILE="main.go"

# --- COLORS FOR OUTPUT ---
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# 1. Validate Arguments
if [ "$#" -ne 2 ]; then
    echo -e "${RED}Error: Missing parameters.${NC}"
    echo -e "Usage: ./run.sh <project_path> <language>"
    echo -e "Example: ./run.sh ./my-java-project java"
    exit 1
fi

PROJECT_PATH=$1
LANG=$2

# 2. Build the program (ensure we have the latest changes)
echo -e "${BLUE}🔨 Building ${BINARY_NAME}...${NC}"
go build -o $BINARY_NAME $MAIN_FILE

if [ $? -ne 0 ]; then
    echo -e "${RED}❌ Build failed! Check your Go code.${NC}"
    exit 1
fi

# 3. Run the program with the provided flags
echo -e "${BLUE}Launching analyzer on $PROJECT_PATH ($LANG)...${NC}"
echo "--------------------------------------------------"

./$BINARY_NAME -path="$PROJECT_PATH" -lang="$LANG"

# 4. Capture exit code of the program
EXIT_CODE=$?
echo "--------------------------------------------------"
if [ $EXIT_CODE -eq 0 ]; then
    echo -e "${BLUE}Program exited cleanly.${NC}"
else
    echo -e "${RED}Program exited with error code: $EXIT_CODE${NC}"
fi