# Windows Remote Control

`veilkey-installer` 기준 Windows 테스트는 `Windows 11 IoT Enterprise LTSC Evaluation` seed VM을 기준으로 진행합니다.

현재 기준 VM:

- seed: `proxmox-test-seed-windows-iot-ltsc-eval`
- 목표 템플릿: `proxmox-test-template-windows-iot-ltsc-eval`
- 목표 랩: `proxmox-test-lab-windows`

## 표준 원격 제어 표면

설치 후 원격 제어는 아래 조합을 기준으로 합니다.

- `RDP`
- `WinRM`
- `QEMU guest agent`

역할:

- `RDP`: GUI 조작
- `WinRM`: CLI/자동화 실행
- `guest agent`: IP 확인, 상태 점검, 파일 주입 보조

## 설치 후 해야 할 일

1. Windows 설치 완료
2. VirtIO 드라이버 설치
3. 네트워크 정상화
4. 관리자 계정/암호 설정
5. Remote Desktop 활성화
6. WinRM 활성화
7. QEMU guest agent 설치

## 연결 확인 기준

최소 기준:

- RDP 포트 열림
- WinRM 포트 열림
- guest agent 응답

기본 포트:

- `3389/tcp` for RDP
- `5985/tcp` for WinRM HTTP
- `5986/tcp` for WinRM HTTPS

## 호스트에서 확인

helper 사용:

```bash
./scripts/windows_remote_check.sh 10.50.x.y
```

직접 확인:

```bash
nc -zv 10.50.x.y 3389
nc -zv 10.50.x.y 5985
nc -zv 10.50.x.y 5986
```

## 템플릿화 기준

아래가 되면 seed를 템플릿으로 올립니다.

- 네트워크 연결 정상
- VirtIO 드라이버 설치 완료
- RDP 접속 성공
- WinRM 연결 성공
- guest agent 응답 성공

그 다음:

- seed -> `proxmox-test-template-windows-iot-ltsc-eval`
- clone -> `proxmox-test-lab-windows`

## 비고

- 설치 전 단계는 Proxmox 콘솔에 의존합니다.
- Windows ISO는 Microsoft Evaluation Center 평가판을 사용합니다.
- 설치 전에는 원격 제어 표면이 없습니다.
