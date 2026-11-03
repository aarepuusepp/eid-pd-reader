# Reader 2.1.11

Adds `--check-pkcs11` for the EID-PD online/offline installer. This command reads
machine configuration and reports the selected module path and readiness as JSON.
It does not connect to a card, read personal data or request a PIN.

Exit 0 means a loadable x64 PKCS#11 DLL exporting C_GetFunctionList; 10 means
missing, 11 wrong architecture, 12 invalid module and 20 invalid explicit
administrator configuration. Readiness does not establish card compatibility
or successful decryption. Contact-card crypto still uses PKCS#11; the supported
direct NFC/PACE path remains independent of it.

Installer 2.1.11 can provision the separate OpenSC package without DigiDoc4.
The existing configuration schema and native messaging protocol are unchanged.
