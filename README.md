## podDeployer-operator 
### pod按照容器权重顺序启动operator
![]()
### 项目思路与设计
设计背景：
集群在部署服务时，可能会需要多个container部署在同一个pod中的场景。当不同的container有顺序需要依赖时，可以采用k8s 提供的hook来做命令脚本

思路：对deployment使用clientSet patch操作实现。

### 项目功能

```yaml
apiVersion: api.practice.com/v1alpha1
kind: Poddeployer
metadata:
  name: mypoddeployer
  namespace: default
spec:
  deployment_spec:
    replicas: 1
    template:
      spec:
        containers:
          - name: example1
            image: busybox:1.34
            command:
              - "sleep"
              - "3600"
          - name: example2
            image: nginx:1.14.2
            ports:
            - containerPort: 80
          - name: example3
            image: nginx:1.14.2
            ports:
            - containerPort: 81
          - name: example4
            image: nginx:1.18-alpine
            ports:
              - containerPort: 82
  priority_images:        # image的权重排序
    - image: example1
      value: 200          # 值越大权重越高
    - image: example2
      value: 50
    - image: example3
      value: 100
    - image: example4
      value: 1000

```