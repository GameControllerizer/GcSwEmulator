# GcSwEmulator
GcSwemulator は，ディジタルゲーム拡張のためのミドルウェア [GameControllerizer](https://github.com/nobu-e753/GameControllerizer) の補助機能であり，
各ゲームプラットフォームに対する入力を電子的に模擬するS/Wです．
起動時に指定したホスト（Node-REDサーバーを想定）より[DSL4GC](https://github.com/nobu-e753/GameControllerizer/blob/master/dsl4gc/README.md)フォーマットの
制御信号を受け取とり，これを再生します．

Read this in other languages: [English](./README.en.md), [日本語](./README.md)

# 利用方法

`bin`フォルダより対象プラットフォームのバイナリをダウンロードしコマンドラインで次のように起動してください．
対象プラットフォームのバイナリが存在しない場合は，下記に示す環境にてビルドを行ってください．

```
% GcSwEmulator_xxxxxx.exe -h
```

# 動作条件
現時点でエミュレート可能なデバイスは以下です．
- PC(Mouse)
- PC(Keyboard)

Go言語で実装されており，ビルドし直すことで Windows/Mac/Linux の各プラットフォームで動作します．
現時点で実績のある動作環境は以下です．
- Windows10(64bit)
- macOS Sierra v10.12.6 (64bit)
- Ubuntu 16.04(64bit)
    - 仮想環境上（VirtualBox）ではマウスの動作がおかしくなる現象が確認されています

ビルドには，以下の環境が必要です．
- [GoLang](https://golang.org/)
- [robotgo](https://github.com/go-vgo/robotgo)
- [paho.mqtt.golang](https://github.com/eclipse/paho.mqtt.golang)
- [golang-set](https://github.com/deckarep/golang-set)
