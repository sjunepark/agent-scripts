# Windows cleanup safety boundaries

Use this reference during every Windows-cleanup task. UI labels can vary by
Windows build, edition, locale, and organizational policy; inspect the installed
system and verify version-sensitive syntax through installed help or current
Microsoft documentation before execution.

## Read-only and diagnostic surfaces

Prefer inspection that does not alter configuration or delete data:

- **Task Manager:** inspect Processes, Performance, and Startup apps. Startup
  impact is measured evidence, not permission to disable an entry.
- **Storage:** inspect Settings > System > Storage, Temporary files, Cleanup
  recommendations, and the red low-capacity state. This bar reports low free
  space; it is distinct from a critical drive-health warning.
- **Updates and reliability:** inspect Windows Update history, Reliability
  Monitor (`perfmon /rel`), and only events matching the symptom’s time window.
- **Power and security:** inspect the active plan with
  `powercfg /getactivescheme`; inspect Defender status and history with
  `Get-MpComputerStatus` and `Get-MpThreatDetection` where available.
- **Capacity and provider health:** inspect `Get-Volume` and `Get-PhysicalDisk`.
  Windows’ built-in critical storage warning covers NVM SSDs, not every SATA SSD
  or hard drive. A critical warning means back up first and stop cleanup work;
  the absence of a warning does not prove every device is healthy.

Announce and obtain confirmation before heavier diagnostics even when they are
non-repairing:

- `perfmon /report` collects a System Diagnostics report and consumes resources.
- `sfc /verifyonly` verifies protected system files without repairing them.
- `DISM.exe /Online /Cleanup-Image /CheckHealth` checks recorded corruption;
  `/ScanHealth` performs a fuller scan.
- `DISM.exe /Online /Cleanup-Image /AnalyzeComponentStore` reports actual
  component-store characteristics and whether cleanup is recommended.
- `chkdsk C:` without repair switches reports status, but can be I/O intensive
  and may report transient errors on an active unlocked volume.
- `defrag C: /A /U` analyzes the drive. Analysis does not authorize optimization.

Never interpret a single warning, process sample, or apparent folder size as a
cause. Correlate it with the reported symptom and repeatable resource evidence.

## Every mutation requires an exact preview and confirmation

### Storage

- Prefer Settings > System > Storage > Temporary files or Cleanup
  recommendations. Preview every selected category and its estimated size.
- Use Disk Cleanup as a supported fallback. **Clean up system files** adds
  system categories whose consequences must be reviewed separately.
- Treat Downloads, Recycle Bin, Desktop, personal folders, browser profiles,
  application data, virtual machines, development environments, and local
  backups as user data. Inventory exact targets and obtain explicit deletion
  confirmation.
- Storage Sense changes persistent policy. Preview its schedule, Recycle Bin and
  Downloads age settings, and cloud-content behavior before enabling or running
  it. It normally applies only to the system drive. On Windows 11 22H2 and later,
  configured cloud-content management can make inactive OneDrive files
  online-only after 30 days.
- Before making OneDrive content online-only, verify sync is healthy and preserve
  **Always keep on this device**. Explain that restoring online-only files needs
  network access and cloud availability. Prefer **Free up space** over deletion;
  ordinary deletion in a synced folder normally propagates to OneDrive.
- Delete `Windows.old` only through the supported Storage or cleanup surface,
  after the user confirms that rollback and file recovery are no longer needed.
  This is irreversible and normally removes the time-limited **Go back** option.

### Startup apps, installed apps, power, updates, and scans

- Disable only an exact, measured, nonessential startup entry. Do not infer that
  security, backup or sync, VPN, accessibility, input, audio, graphics, storage,
  or other device-support entries are expendable. Record how to re-enable it.
- Uninstall an exact app only after reviewing local data, dependencies, installer
  availability, license or account requirements, and reversal. Repairing a
  malfunctioning app may be more appropriate than removing it.
- Changing power mode trades energy use, battery life, heat, and fan noise for
  performance. Keep the observed mode unless the user confirms an explained
  change; Balanced is the neutral default.
- Inspect update history before checking for new updates. Checking, downloading,
  installing Windows or optional driver updates, and restarting are distinct
  confirmed actions. Do not install third-party driver-updater software.
- A Defender Quick, Full, Custom, or Offline scan consumes resources and can
  quarantine or remove detections under current policy. Preview scan type and
  likely duration. Offline Scan always restarts into Windows Recovery
  Environment, so require saved work and separate restart confirmation.

### Drives, components, and system repair

- Windows normally optimizes drives on a schedule. Analyze first. If evidence
  justifies a manual run, use media-appropriate optimization through **Defragment
  and Optimize Drives** or `defrag C: /O /U /V`; never force an HDD-specific
  operation onto an SSD merely to “clean” it.
- Prefer Windows’ scheduled component servicing. The supported
  `schtasks.exe /Run /TN "\Microsoft\Windows\Servicing\StartComponentCleanup"`
  task retains superseded components for a grace period but still mutates system
  state and requires confirmation.
- `DISM.exe /Online /Cleanup-Image /StartComponentCleanup` immediately removes
  superseded component versions without the normal scheduled grace period. Use
  only when component analysis justifies it and the user confirms that specific
  consequence.
- Treat `DISM.exe /Online /Cleanup-Image /RestoreHealth` followed by
  `sfc /scannow` as a diagnosed corruption repair, not a generic speed boost.
  Preview elevation, possible Windows Update network use, duration, and recovery
  prerequisites; do not interrupt them.
- Treat `chkdsk /f`, `/r`, `/x`, or `/b` as repair operations that can change the
  volume and require a restart or offline lock. Do not escalate from status-only
  inspection without a separate diagnosis, backup check, and confirmation.

Before nontrivial system repair or a consequential system-setting change,
verify a current personal-file backup and consider System Protection. A restore
point covers system files, apps, registry, and settings—not personal files—and
System Protection may be disabled. Do not purge restore points as routine space
cleanup.

## Prohibited normal-cleanup operations

- Never install, recommend, or run registry-cleaning utilities or bulk registry
  optimization. A narrow registry change belongs to a diagnosed repair workflow,
  with an affected-key export, restore point, and explicit authority.
- Never manually delete, take ownership of, change ACLs on, or replace content
  under `%WINDIR%`, `System32`, `WinSxS`, or `C:\Windows\Installer`. Never delete
  arbitrary content under `ProgramData` or user `AppData` without a documented
  owner-specific lifecycle.
- Never offer DISM `/ResetBase` or `/SPSuperseded` during normal cleanup. They
  remove the ability to uninstall installed updates or service packs.
- Never manually delete or disable `pagefile.sys` as cleanup. Page files support
  the system commit limit and crash dumps.
- Never manually delete `hiberfil.sys`. `powercfg /hibernate off` disables
  hibernation and affects hibernation-dependent Fast Startup; treat it as a
  separate power-feature decision, not generic cleanup.
- Never weaken Defender, firewall, UAC, BitLocker, update policy, or another
  security control to improve performance.
- Never reset or reinstall Windows, repartition a drive, change firmware or TPM
  state, or securely erase a device under this skill.

## Windows 10 lifecycle

Microsoft ended free security updates and support for ordinary Windows 10
editions on October 14, 2025. Before describing the device as unsupported,
identify its edition and servicing channel and inspect whether it is receiving
current updates through an Extended Security Updates program. Some LTSC editions
have their own later lifecycle; for example, Windows 10 Enterprise LTSC 2021 is
listed through January 12, 2027. Report the device's observed status and avoid
claiming that cleanup changes it. When migration is appropriate, recommend a
separately reviewed path to a supported operating system; do not initiate an
upgrade under this skill.

## Primary Microsoft references

Sources were checked on 2026-08-13. Prefer the installed system and current
Microsoft pages when behavior differs.

- [Tips to improve PC performance in Windows](https://support.microsoft.com/en-us/windows/tips-to-improve-pc-performance-in-windows-b3b3ef5b-5953-fb6a-2528-4bbed82fba96)
- [Configure Startup applications in Windows](https://support.microsoft.com/en-US/Windows/Experience/Startup-Boot/configure-startup-applications-in-windows)
- [Manage drive space with Storage Sense](https://support.microsoft.com/en-US/Windows/Experience/Storage-FileManagement/manage-drive-space-with-storage-sense)
- [Free up drive space in Windows](https://support.microsoft.com/en-US/Windows/Experience/Storage-FileManagement/free-up-drive-space-in-windows)
- [What to do about a critical storage warning](https://support.microsoft.com/en-US/Windows/Experience/Storage-FileManagement/what-to-do-about-a-critical-warning-for-a-storage-device)
- [Microsoft policy for registry-cleaning utilities](https://support.microsoft.com/en-us/topic/microsoft-support-policy-for-the-use-of-registry-cleaning-utilities-0485f4df-9520-3691-2461-7b0fd54e8b3a)
- [Clean up the WinSxS folder](https://learn.microsoft.com/en-us/windows-hardware/manufacture/desktop/clean-up-the-winsxs-folder?view=windows-11)
- [Missing Windows Installer cache requires computer rebuild](https://learn.microsoft.com/en-us/troubleshoot/windows-client/application-management/missing-windows-installer-cache)
- [Use System File Checker to repair Windows](https://support.microsoft.com/en-us/windows/experience/backup-recovery/use-the-system-file-checker-tool-to-repair-missing-or-corrupted-system-files)
- [CHKDSK command reference](https://learn.microsoft.com/en-us/windows-server/administration/windows-commands/chkdsk)
- [Defrag command reference](https://learn.microsoft.com/en-us/windows-server/administration/windows-commands/defrag)
- [Virus and threat protection in Windows Security](https://support.microsoft.com/en-us/windows/security/threat-malware-protection/virus-and-threat-protection-in-the-windows-security-app)
- [System Protection](https://support.microsoft.com/en-us/windows/experience/backup-recovery/system-protection)
- [Back up a BitLocker recovery key](https://support.microsoft.com/en-US/Windows/Security/encryption/back-up-your-bitlocker-recovery-key)
- [Windows 10 support ended on October 14, 2025](https://support.microsoft.com/topic/4fdf8a9e-ddc9-4f65-971f-47e7debab6e1)
- [Windows 10 Extended Security Updates program](https://learn.microsoft.com/en-us/windows/whats-new/extended-security-updates)
- [Windows 10 Enterprise LTSC 2021 lifecycle](https://learn.microsoft.com/en-us/lifecycle/products/windows-10-enterprise-ltsc-2021)
