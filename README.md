# go-asmr-spider

简单的音声下载器。

## 怎么用

开箱即用, 无需额外配置环境。

1. 前往 [actions](https://github.com/DiheChen/go-asmr-spider/actions) 下载与你操作系统 / 平台对应的可执行文件, 运行一次, 生成默认的配置文件。

2. 再次运行程序, 输入你要下载的音声的 RJ 号, 如果要下载多个音声用空格隔开。

由于众所周知的问题, Windows 的文件系统不支持 `?`, `/`, `\`, `*`, `:`, `|` 等字符作为文件名。

如果在 Windows 下运行本程序且音声的文件名出现了这些字符, 会被替换成 `_`。

还有个 [Python 分支](https://github.com/DiheChen/go-asmr-spider/tree/python), 因为部署很麻烦所以不推荐。

## 致谢

感谢 <https://asmr.one>, 现在每天都有不同的女孩子陪我睡觉。