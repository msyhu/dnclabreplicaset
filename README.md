# dnclabreplicaset
## 목표
Pod 개수를 계산해주는 특정 공식 or AI 기법이 구현되어있는 파이썬 서버에 주기적으로 Pod 개수를 질의한 후, 그 개수에 맞추어 Pod 개수를 조절하는 custom controller 를 구현한다.

![image](https://user-images.githubusercontent.com/81010357/125193006-5769a080-e285-11eb-8487-5aa623d8de85.png)
![image](https://user-images.githubusercontent.com/81010357/125193035-6e0ff780-e285-11eb-99ec-6c014cfa4404.png)
![image](https://user-images.githubusercontent.com/81010357/125193047-7d8f4080-e285-11eb-86f6-cf8cd2b518b2.png)
![image](https://user-images.githubusercontent.com/81010357/125193063-95ff5b00-e285-11eb-96b1-ade9543bca2e.png)

## 사용한 명령어들
- make install
- make run
- kubebuilder init --domain ds.korea.ac.kr
- kubebuilder create api --group mycore --version v1 --kind DnclabReplicaSet

## reference
- https://if.kakao.com/session/101
