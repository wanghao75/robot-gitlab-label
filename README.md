# robot-gitee-label
[中文README](README_zh_CN.md)

### Overview

This is the gitee code hosting platform label processing robot, code cloud platform users can use a number of commands to add and remove issue and Pull Request labels; in addition will listen to the PR change event, according to the configuration of the automatic removal of the specified label and test whether the label is valid.

### Function

- **Command**

  | command            | example                                  | describe                                                     | who can use                                                  |
  | ------------------ | ---------------------------------------- | ------------------------------------------------------------ | ------------------------------------------------------------ |
  | /[remove-]kind     | /kind bug<br/>/remove-kind bug           | Add or remove this kind of kind type label. Example: `kind/bug` label. | Anyone can trigger such a command on a Pull Request or Issue. |
  | /[remove-]priority | /priority high<br/>/remove-priority high | Add or remove this kind of priority type label. Example: `priority/high` label. | Anyone can trigger such a command on a Pull Request or Issue. |
  | /[remove-]sig      | /sig kernel<br/>/remove-sig kernel       | Add or remove this kind of sig type label. Example: `sig/kernel`label。 | Anyone can trigger such a command on a Pull Request or Issue. |

  **Note: To prevent repository label controllable, only the repository collaborators can use the command to add non-existent labels for the repository (that is, create new labels), non-repository collaborators will prompt a tagging failure**

- **Clean up labels**

  When PR has new commits, those labels that need to be cleared will be automatically removed according to the configuration.

- **Verify the timeliness of the label**

  Support setting the time limit of label, when PR occurs edit event, it will trigger to detect whether the label has been invalidated according to the configuration and remove the invalidated label.

### Configuration

example

````yaml
config_items:
  - repos:  #List of repositories to be managed by robot
     -  owner/repo
     -  owner1
    excluded_repos: #Robot manages the list of repositories to be excluded
     - owner1/repo1
    clear_labels: # List of labels that need to be removed after a source branch changed event
     - lgtm
     - approve
    labels_to_validate: #Verify the label's time-sensitive configuration
      - label: ci-pipline-success
        active_time: 1 #Indicates how many hours after the creation of the label to expire
````

