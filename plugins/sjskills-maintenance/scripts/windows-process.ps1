$ErrorActionPreference = 'Stop'
# Assign the supervisor before launching anything. Descendants inherit the job,
# including when the native parent exits before its children close their pipes.
Add-Type -TypeDefinition @'
using System;
using System.ComponentModel;
using System.Runtime.InteropServices;
public static class StartupProcessJob {
    [StructLayout(LayoutKind.Sequential)] struct Basic {
        public long ProcessTime, JobTime;
        public uint Flags;
        public UIntPtr MinimumWorkingSet, MaximumWorkingSet;
        public uint ActiveProcesses;
        public UIntPtr Affinity;
        public uint Priority, Scheduling;
    }
    [StructLayout(LayoutKind.Sequential)] struct Extended {
        public Basic Basic;
        public ulong ReadOperations, WriteOperations, OtherOperations;
        public ulong ReadBytes, WriteBytes, OtherBytes;
        public UIntPtr ProcessMemory, JobMemory, PeakProcessMemory, PeakJobMemory;
    }
    [DllImport("kernel32.dll", SetLastError=true)] static extern IntPtr CreateJobObject(IntPtr attributes, string name);
    [DllImport("kernel32.dll", SetLastError=true)] static extern bool SetInformationJobObject(IntPtr job, int kind, ref Extended info, uint size);
    [DllImport("kernel32.dll", SetLastError=true)] static extern bool AssignProcessToJobObject(IntPtr job, IntPtr process);
    [DllImport("kernel32.dll")] static extern IntPtr GetCurrentProcess();
    // Deliberately retained until process exit. The OS closes the final handle
    // even when Node forcibly terminates this supervisor.
    static IntPtr job;
    public static void OwnTree() {
        job = CreateJobObject(IntPtr.Zero, null);
        if (job == IntPtr.Zero) throw new Win32Exception();
        Extended info = new Extended();
        info.Basic.Flags = 0x2000; // JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE
        if (!SetInformationJobObject(job, 9, ref info, (uint)Marshal.SizeOf(info)) ||
            !AssignProcessToJobObject(job, GetCurrentProcess())) throw new Win32Exception();
    }
}
'@
[StartupProcessJob]::OwnTree()
$invocation = $env:SJSKILLS_CHECK_PROCESS | ConvertFrom-Json
Remove-Item Env:SJSKILLS_CHECK_PROCESS
[Console]::OutputEncoding = [System.Text.UTF8Encoding]::new($false)
$OutputEncoding = [Console]::OutputEncoding
$executable = $invocation.executable
$arguments = @($invocation.arguments)
& $executable @arguments
exit $LASTEXITCODE
