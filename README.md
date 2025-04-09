# KubeSphere MCP Server
get resource from KubeSphere. Divided into four modules: Workspace Management, Cluster Management, User and Roles, Extensions Center.

# Use by stdio
## build binary 
```shell
go build -o ks-mcp-server cmd/main.go
```

## create ksconfig to access KubeSphere
for example:
```yaml
apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: xxxx
    server: https://kubesphere.com
  name: kubesphere
contexts:
- context:
    cluster: kubesphere
    user: admin
  name: admin@kubesphere
current-context: admin@kubesphere
kind: Config
preferences: {}
users:
- name: admin
  username: xxxx
  password: xxxx
  user:
    client-certificate-data: xxxx
    client-key-data: xxx

```

## configuration mcpServer in Claude Desktop
```json
{
  "mcpServers": {
    "KS-MCP-Server": {
      "args": [
        "stdio",
        "--ksconfig", "xxx/kubeconfig",
        "--ks-apiserver", "http://172.0.0.1:30881"
      ],
      "command": "ks-mcp-server"
    }
  }
}
```

# use by sse
## create ksconfig to access KubeSphere
for example:
```yaml
apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: xxxx
    server: https://kubesphere.com
  name: kubesphere
contexts:
- context:
    cluster: kubesphere
    user: admin
  name: admin@kubesphere
current-context: admin@kubesphere
kind: Config
preferences: {}
users:
- name: admin
  username: xxxx
  password: xxxx
  user:
    client-certificate-data: xxxx
    client-key-data: xxx

```

## run sse server
```shell
go run cmd/main.go sse --ksconfig xxx/ksconfig
```
## 
it will serve in 8080 port.

## configuration mcpServer in Claude Desktop
```json
{
  "mcpServers": {
    "KS-MCP-Server": {
      "args": [
        "-y",
        "mcp-remote",
        "http://localhost:8080/sse",
        "--clean"
      ],
      "command": "npx"
    }
  }
}
```

# run in Claude Desktop
![execute-result](image.png)