# protoc

### Compile `.proto` files
```bash
protoc -I=./proto --go_out=./proto ./proto/id.proto ./proto/employee.proto
```

### Generate descriptor set
```bash
protoc -I=./proto --descriptor_set_out example.ds ./proto/id.proto ./proto/employee.proto
```
