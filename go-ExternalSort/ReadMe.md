# 100GB数据进行外部排序
## 思路

![流程图](image/pipeline.png)

## 性能:
没有buffer的情况
![没有buffer的情况](image/1.png)

有buffer的情况，节约了写入chan的切换时间
![有buffer的情况](image/2.png)

开启更多协程的情况
![开启更多协程的情况](image/3.png)

来源：
https://www.imooc.com/video/16273