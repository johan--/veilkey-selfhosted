# Android Remote Control

`veilkey-installer` 기준 모바일 테스트는 Android만 지원합니다.

현재 기준 VM:

- seed: `proxmox-test-seed-android`
- 목표 템플릿: `proxmox-test-template-android`
- 목표 랩: `proxmox-test-lab-android`

## 표준 원격 제어 표면

설치 후 원격 제어는 `ADB over TCP`를 기준으로 합니다.

이유:

- Proxmox 콘솔에 의존하지 않고 외부에서 조종 가능
- Android x86 / BlissOS에서 가장 단순한 자동화 표면
- 이후 `scrcpy`, `adb shell`, APK 설치, UI 자동화로 확장 가능

## 설치 후 해야 할 일

BlissOS 설치가 끝나면 아래를 수행합니다.

1. 개발자 옵션 활성화
2. USB debugging 활성화
3. Wireless debugging 또는 TCP ADB 활성화
4. Android IP 확인
5. 호스트에서 `adb connect <ip>:5555`

## 호스트에서 연결

helper 사용:

```bash
./scripts/android_adb_connect.sh 10.50.x.y
```

직접 사용:

```bash
adb connect 10.50.x.y:5555
adb devices
adb shell getprop ro.product.model
```

## 연결 확인 기준

최소 확인:

```bash
adb devices
adb shell getprop ro.build.version.release
adb shell settings get global adb_enabled
```

## 템플릿화 기준

아래가 되면 seed를 템플릿으로 올립니다.

- `adb connect` 성공
- `adb shell` 성공
- IP가 고정 또는 추적 가능
- 기본 앱/초기 설정 완료

그 다음:

- seed -> `proxmox-test-template-android`
- clone -> `proxmox-test-lab-android`

## 비고

- Android는 현재 guest agent를 기대하지 않습니다.
- 설치 전 단계는 Proxmox 콘솔로만 진행합니다.
- 설치 후 운영 표면은 `ADB over TCP` 하나로 통일합니다.
