## kusion destroy

删除运行时中指定的资源

### Synopsis

通过资源规约删除资源。

 只接受 KCL 文件。只能指定一种类型的参数：文件名、资源、名称、资源或标签选择器。

 请注意，destroy 命令不会进行资源版本检查， 因此如果有人在你提交销毁时提交了对资源的更新， 他们的更新将与资源一起丢失。

```
kusion destroy [flags]
```

### Examples

```
  # 删除 pod.k 中的配置
  kusion destroy -f ./pod.k
```

### Options

```
  -D, --argument stringToString   指定顶级参数 (default [])
  -C, --backend-config strings    backend-config 配置状态存储后端信息
      --backend-type string       backend-type 指定状态存储后端类型，支持 local、db、oss 和 s3
  -d, --detail                    预览后自动展示 apply 计划细节
  -h, --help                      help for destroy
      --operator string           指定操作人
  -O, --overrides strings         指定配置覆盖路径和值
  -Y, --setting strings           指定命令行配置文件
  -w, --workdir string            指定工作目录
  -y, --yes                       预览后自动审批并应用更新
```

### SEE ALSO

* [kusion](kusion.md)	 - Kusion 是 KusionStack 的平台工程引擎

###### Auto generated by spf13/cobra on 28-Sep-2023
