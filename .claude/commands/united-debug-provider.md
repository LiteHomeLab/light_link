框架需要进行 provider 功能的测试和验收
 - NATS 服务器在 172.18.200.47:4222 已经启动，使用了私有证书；
 - 在 \light_link_platform\client 目录下有对应 NATS 的私有证书，不要自己单独生成；
 - 在后台启动 @light_link_platform\examples\provider\ 下面所有的服务提供者；
 - 在后台启动：管理平台 manager_base  @light_link_platform\manager_base\ （manager_base包含前、后端）；