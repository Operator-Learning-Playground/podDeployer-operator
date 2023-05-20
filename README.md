## podDeployer-operator 按照容器权重顺序启动 operator
![]()
### 有重大 bug! patch 不能保证容器顺序启动

### 项目思路与设计
设计背景：
集群在部署服务时，可能会需要多个 container 部署在同一个 pod 中的场景。当不同的 container 有顺序需要依赖时，可以采用 k8s 提供的 hook 来做命令脚本，也能使用 patch 进行操作。

思路：对 deployment daemonset statefulset 使用 clientSet patch 操作实现。

### 项目功能
- 使用 patch 操作实现容器顺序启动功能(仅仅用于模拟)。 可参考 [deployment yaml参考](yaml/example_deployment.yaml) [statefulset yaml参考](yaml/example_statefulset.yaml) [daemonset yaml参考](yaml/example_daemonset.yaml)
```yaml
apiVersion: api.practice.com/v1alpha1
kind: Poddeployer
metadata:
  name: mypoddeployer-deployment
  namespace: default
spec:
  type: apps/v1/deployments # 可选择使用哪种类型，目前支持三种：apps/v1/deployments  apps/v1/statefulsets apps/v1/daemonsets
  # 资源对象的原生 Spec
  # 目前支持三种：deployment_spec statefulset_spec daemonset_spec
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