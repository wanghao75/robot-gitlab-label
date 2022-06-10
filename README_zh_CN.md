# robot-gitee-label

### 概述

这是码云代码托管平台标签处理机器人，码云平台的用户可以使用一些命令添加和移除issue与PullRequest标签；此外还会监听PR改变的事件，根据配置自动移除指定的标签以及检测标签是否有效。

### 功能

- **命令**

  该机器人提供的指令如下表所示。

  | 命令               | 示例                                     | 描述                                                         | 谁能使用                                              |
  | ------------------ | ---------------------------------------- | ------------------------------------------------------------ | ----------------------------------------------------- |
  | /[remove-]kind     | /kind bug<br/>/remove-kind bug           | 添加或者删除这种kind类型的标签。 例如：`kind/bug`标签。      | 任何人都能在一个Pull Request或者Issue上触发这种命令。 |
  | /[remove-]priority | /priority high<br/>/remove-priority high | 添加或者删除这种priority类型的标签。 例如：`priority/high`标签。 | 任何人都能在一个Pull Request或者Issue上触发这种命令。 |
  | /[remove-]sig      | /sig kernel<br/>/remove-sig kernel       | 添加或者删除这种sig类型的标签。 例如：`sig/kernel`标签。     | 任何人都能在一个Pull Request或者Issue上触发这种命令。 |

  **注意：为防止仓库标签可控，只有仓库的协作者可以使用指令为仓库打上不存在的标签（也就是创建新的标签），非仓库协作者将会提示打标签失败**

- **清理标签**

  当PR有新的commit，根据配置将自动删除这些配置标签。

- **验证标签的时效性**

  支持设置标签的时效，当PR发生编辑事件，就会触发根据配置检测标签是否已经失效并移除失效的标签。

### 配置

例子：

```yaml
config_items:
  - repos:  #robot需管理的仓库列表
     -  owner/repo
     -  owner1
    excluded_repos: #robot 管理列表中需排除的仓库
     - owner1/repo1
    clear_labels: # source branch changed 事件发生后 需要移除的标签列表
     - lgtm
     - approve
    labels_to_validate: #验证标签时效性的配置
      - label: ci-pipline-success
        active_time: 1 #表示标签创建后多少小时失效
```



