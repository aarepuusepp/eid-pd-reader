# Risk Register — WebEID Personal Data Fallback

## Risk Assessment Matrix

| Risk ID | Description | Probability | Impact | Severity | Mitigation | Status |
|---------|-------------|------------|--------|----------|------------|--------|
| R1 | PC/SC service unavailable | Medium | High | High | Retry logic, clear errors | Mitigated |
| R2 | Card reader driver issues | Medium | Medium | Medium | Document requirements, test matrix | Mitigated |
| R3 | Extension installation friction | High | Medium | Medium | Clear docs, install scripts | Mitigated |
| R4 | Host manifest misconfiguration | Medium | High | High | Automated install scripts, validation | Mitigated |
| R5 | Different card profiles (EF 5044) | Low | Medium | Low | FID overrides, documentation | Accepted |
| R6 | NFC complexity (PACE/CAN) | Low | Low | Low | Isolated design, future work | Deferred |
| R7 | Security: Origin spoofing | Low | High | High | Origin pinning, manifest restrictions | Mitigated |
| R8 | Performance degradation | Medium | Medium | Medium | Timeouts, monitoring, rollback | Mitigated |
| R9 | User confusion (cert vs fallback) | Medium | Low | Low | Clear UX, source indicator | Mitigated |
| R10 | Data exposure (logging) | Low | High | High | Logging policy, code review | Mitigated |

## Detailed Risk Analysis

### R1: PC/SC Service Unavailable

**Description:**
PC/SC service (SCardSvr on Windows, pcscd on Linux) not running or unavailable.

**Impact:**
- Native helper cannot connect to card readers
- All fallback attempts fail
- Users see errors

**Probability:** Medium (common on fresh installs, service restarts)

**Mitigation:**
- Clear error messages: "PC/SC service not running"
- Documentation: Check service status before use
- Retry logic: Automatic retry with backoff
- User guidance: How to start service

**Status:** Mitigated

---

### R2: Card Reader Driver Issues

**Description:**
Card reader drivers not installed or incompatible with PC/SC.

**Impact:**
- Readers not detected
- Connection failures
- Platform-specific issues

**Probability:** Medium (varies by OS, reader model)

**Mitigation:**
- Document supported readers
- Test matrix: Windows/Linux/macOS
- Clear error messages
- Link to driver downloads

**Status:** Mitigated

---

### R3: Extension Installation Friction

**Description:**
Users struggle to install extension or configure host manifests.

**Impact:**
- Low adoption
- Support tickets
- Feature not used

**Probability:** High (manual installation required)

**Mitigation:**
- Step-by-step installation guide
- Automated install scripts (`ops/install-host.ps1`, `ops/install-host.sh`)
- Video tutorial (optional)
- Support team training

**Status:** Mitigated

---

### R4: Host Manifest Misconfiguration

**Description:**
Host manifest path incorrect, extension ID mismatch, or registry/file permissions.

**Impact:**
- Native Messaging fails
- Feature completely broken
- Silent failures

**Probability:** Medium (common during setup)

**Mitigation:**
- Automated install scripts validate paths
- Absolute paths required
- Extension ID validation
- Clear error messages in extension

**Status:** Mitigated

---

### R5: Different Card Profiles (EF 5044)

**Description:**
Older Estonian ID-cards use record file EF 5044 instead of transparent EFs 5001-500F.

**Impact:**
- PD read fails for older cards
- Users with older cards cannot use fallback

**Probability:** Low (most cards are ID1 profile)

**Mitigation:**
- FID override support (allows custom FIDs)
- Documentation: How to extend for EF 5044
- Future enhancement: Auto-detect card profile

**Status:** Accepted (low priority)

---

### R6: NFC Complexity (PACE/CAN)

**Description:**
NFC requires PACE handshake with CAN, more complex than contact interface.

**Impact:**
- NFC not supported initially
- Users with NFC-only readers cannot use feature

**Probability:** N/A (out of scope for M1-M4)

**Mitigation:**
- Isolated backend interface design
- NFC implementation deferred
- Document as future enhancement

**Status:** Deferred

---

### R7: Security: Origin Spoofing

**Description:**
Malicious website attempts to read PD by impersonating legitimate app.

**Impact:**
- Unauthorized PD access
- Privacy violation
- Compliance issues

**Probability:** Low (mitigations in place)

**Mitigation:**
- Origin pinning in `manifest.json` (`externally_connectable.matches`)
- Host manifest `allowed_origins` restricted to extension ID
- Content script validates `event.source === window`
- No wildcards in production

**Status:** Mitigated

---

### R8: Performance Degradation

**Description:**
Fallback path slower than cert path, causing user frustration.

**Impact:**
- Poor user experience
- Timeout issues
- Increased support tickets

**Probability:** Medium (fallback involves more steps)

**Mitigation:**
- Configurable timeout (`PD_FALLBACK_TIMEOUT_MS`)
- Monitor response times
- Optimize native helper (already fast)
- Rollback if performance unacceptable

**Status:** Mitigated

---

### R9: User Confusion (Cert vs Fallback)

**Description:**
Users don't understand when fallback is used vs cert path.

**Impact:**
- Support questions
- User frustration
- Trust issues

**Probability:** Medium (new concept)

**Mitigation:**
- `source` field in response (`"cert"` vs `"fallback:chip"`)
- Clear UI messages (optional)
- Documentation explains flow
- Support team training

**Status:** Mitigated

---

### R10: Data Exposure (Logging)

**Description:**
PD values accidentally logged or exposed in telemetry.

**Impact:**
- Privacy violation
- GDPR compliance issue
- Security incident

**Probability:** Low (policy in place)

**Mitigation:**
- Logging policy: Status only, no PII
- Code review checklist
- Telemetry: Truncated error messages only
- Security audit before deployment

**Status:** Mitigated

---

## Risk Monitoring

### Ongoing Monitoring

- **Weekly**: Review error rates and telemetry
- **Monthly**: Security audit of logs
- **Quarterly**: Risk register review and update

### Escalation Criteria

- **High Severity**: Immediate rollback, security team notification
- **Medium Severity**: Investigate within 24 hours, consider rollback
- **Low Severity**: Track and address in next sprint

## Risk Response Plan

### If Risk Materializes

1. **Assess**: Determine actual impact
2. **Contain**: Disable feature flag if needed
3. **Investigate**: Root cause analysis
4. **Remediate**: Fix issue
5. **Verify**: Test fix thoroughly
6. **Communicate**: Notify stakeholders
7. **Document**: Update risk register

## Residual Risks

After mitigations, remaining risks:

- **R5**: Older card profiles (accepted, low priority)
- **R6**: NFC support (deferred, future work)
- **R9**: User confusion (mitigated but may need UX improvements)

## Risk Owner

- **Overall**: Tech Lead / Product Owner
- **Security Risks**: Security Team
- **Performance Risks**: Engineering Team
- **UX Risks**: Product Team

