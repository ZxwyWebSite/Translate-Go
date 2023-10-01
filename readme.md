## ZxwyWebSite/Translate-Go
### 简介
+ Zxwy翻译(生草)机的后端
+ 测试版，部分功能暂未完善

### 使用
+ 构建后直接运行即可，默认监听1009端口

### 构建
+ 前端构建后进入build目录执行 `zip statics.zip -r *` 打包静态文件
+ 将statics.zip移动到此项目，替换原文件
+ 正常构建即可

### 问题
+ 前端src/App.tsx(90:133)的encodeURI()无法转义";"导致text参数解析错误，应改为encodeURIComponent()
+ go-googletrans库translate.go(126:16)返回nil结构体导致获取Text时发生Panic

### WS接口
+ (测试版，可能会更改格式，暂未在前端兼容)
+ 地址: ws://192.168.10.22:1009/api/ws (自行替换ip)
+ 支持 trans: 翻译模式 & grass: 生草模式
+ 暂未添加chknil验证
+ 发送示例
    ```json
    {
        "type": "grass",
        "data": {
            "from": "de",
            "to": "zh-CN",
            "text": "Wein und Gesang, Lebensgeometrie! Zum Beispiel hat der Tau am Morgen am vergangenen Tag viel Bitterkeit.\nGroßzügigkeit sollte großzügig sein und Sorgen und Gedanken bleiben unvergesslich. Warum ärgern? Nur Du Kang.\nGrüner Zijin, Youyouwoxin. Aber dem König zuliebe habe ich bisher nachgedacht.\nYo Yo Lu Ming, der Apfel der Wildnahrung. Ich habe Gäste, Harfenblasender Sheng.",
            "force": 3
        }
    }
    ```
+ 返回示例
    ```json
    {
      "type": "stat",
      "msg": "de->pl",
      "data": {
        "num": 1,
        "text": "Wino i piosenka, geometria życia! Na przykład rosa o poranku poprzedniego dnia ma dużo goryczy.\nHojność powinna być hojna, a zmartwienia i myśli pozostaną niezapomniane. Po co się złościć? Tylko ty, Kangu.\nZielony Zijin, Youyouwoxin. Ale na litość króla, myślałem o tym do tej pory.\nYo Yo Lu Ming, dzikie jabłko. Mam gości, Sheng grający na harfie."
      }
    }

    {
      "type": "stat",
      "msg": "pl->ca",
      "data": {
        "num": 2,
        "text": "Vi i cançó, la geometria de la vida! Per exemple, la rosada del matí anterior té molta amargor.\nLa generositat ha de ser generosa i les preocupacions i els pensaments romandran inoblidables. Per què enfadar-se? Només tu, Kang.\nGreen Zijin, Youyouwoxin. Però pel bé del rei, hi he estat pensant fins ara.\nYo Yo Lu Ming, poma salvatge. Tinc convidats, Sheng tocant l'arpa."
      }
    }

    {
      "type": "stat",
      "msg": "ca->ko",
      "data": {
        "num": 3,
        "text": "와인과 노래, 삶의 기하학! 예를 들어, 전날 아침 이슬에는 쓴맛이 많이 있습니다.\n관대함은 관대해야 하며 관심과 생각은 잊혀지지 않을 것입니다. 왜 화를 내나요? 너뿐이야, 강.\n그린진, 유유웍신. 하지만 왕을 위해서 나는 지금까지 그런 생각을 하고 있었습니다.\n요요루밍(Yo Yo Lu Ming), 야생 사과. 손님이 있는데, 하프를 연주하는 Sheng입니다."
      }
    }

    {
      "type": "callback",
      "msg": "de->pl->ca->ko->zh-CN",
      "data": {
        "time": "11.030811353s",
        "text": "美酒，歌曲，还有生命的几何！例如，前一天早上的露水里有很多苦味。\n慷慨一定要慷慨，你的关心和思念不会被忘记。你为什么生气？就是你，康。\n绿色牛仔裤，You Yu Wok Shin。但为了国王，我到现在为止一直这么想。\n哟哟路明，野苹果。我们有一位客人，盛，他弹竖琴。"
      }
    }
    ```
+ 调试工具: https://websocketking.com/

### 更新
#### 2023-10-01
+ 上传源码
