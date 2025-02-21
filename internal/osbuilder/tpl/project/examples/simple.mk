# 变量定义  
APP_NAME = myapp  
SRC = ./...  
OUTPUT_DIR = ./bin  
BUILD_FILE = $(OUTPUT_DIR)/$(APP_NAME)  

# 默认目标  
all: build  

# 构建二进制文件  
build:  
	@echo "Building the application..."  
	@mkdir -p $(OUTPUT_DIR)  
	go build -o $(BUILD_FILE) ./main.go  

# 运行应用程序  
run:  
	@echo "Running the application..."  
	@$(BUILD_FILE)  

# 运行测试  
test:  
	@echo "Running tests..."  
	go test $(SRC)  

# 格式化代码  
fmt:  
	@echo "Formatting code..."  
	go fmt $(SRC)  

# 清理生成的文件  
clean:  
	@echo "Cleaning build files..."  
	@rm -rf $(OUTPUT_DIR)  

# 伪目标（防止文件与目标名称冲突）  
.PHONY: all build run test fmt clean
