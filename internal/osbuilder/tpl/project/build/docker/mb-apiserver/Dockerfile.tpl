# Define the base image for the production environment as Debian Trixie
ARG prod_image=debian:trixie

# Use the specified base image
FROM ${prod_image}
# Set maintainer information
LABEL maintainer="<colin404@foxmail.com>"

# Set the working directory to /opt/{{.ProjectName}}
WORKDIR /opt/{{.ProjectName}}

RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
      echo "Asia/Shanghai" > /etc/timezone  # Set the timezone to Shanghai

RUN mkdir -p /opt/{{.ProjectName}}/log # Create a log directory

# Copy the {{.BinaryName}} executable file to the bin directory in the working directory
COPY {{.BinaryName}} /opt/{{.ProjectName}}/bin/

# Specify the command to be executed when the container starts
ENTRYPOINT ["/opt/{{.ProjectName}}/bin/{{.BinaryName}}"]
