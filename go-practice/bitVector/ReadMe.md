golang练手小项目系列(1)-位向量

本系列整理了10个工作量和难度适中的Golang小项目，适合已经掌握Go语法的工程师进一步熟练语法和常用库的用法。

问题描述：
在数据流分析领域，集合元素都是小的非负整型，集合拥有许多元素，而且集合的操作多数是求并集和交集，位向量是个理想的数据结构。
有一组非负整数，实现一个位向量类型，能在O(1)时间内完成插入、删除和查找等操作。

要点：

实现Has(uint)、Add(uint)、Remove(uint)、Clear()、Copy()、String()、AddAll(…uint)、UnionWith()、IntersectWith()、DifferenceWith()方法。


拓展：
实现uint64 uint32 uint三种版本