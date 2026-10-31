> Historical development record. Use [current reader README](../../README.md) for supported setup; these notes are not current installation/API instructions.

# Rollout Plan — WebEID Personal Data Fallback

## Overview

This document outlines the phased rollout strategy for the WebEID Personal Data fallback feature, using feature flags and gradual enablement to minimize risk.

## Rollout Phases

### Phase 0: Pre-Deployment (Week 1)

**Objectives:**
- Complete development and testing
- Prepare deployment artifacts
- Set up monitoring/telemetry

**Tasks:**
- [ ] Build native helper binaries for all platforms
- [ ] Package extension
- [ ] Create installation documentation
- [ ] Set up telemetry endpoints (status only, no PII)
- [ ] Internal QA with Thales Estonian ID-card

**Success Criteria:**
- All QA checklist items pass
- Build artifacts available
- Documentation complete

### Phase 1: Internal Testing (Week 2)

**Objectives:**
- Validate with internal team
- Gather initial feedback
- Fix critical issues

**Configuration:**
```javascript
PD_FALLBACK_ENABLED = false  // Cert-only mode
```

**Tasks:**
- [ ] Deploy code with feature flag disabled
- [ ] Internal team tests cert path (baseline)
- [ ] Monitor error rates
- [ ] Document any issues

**Success Criteria:**
- No regressions in cert path
- Error rates stable
- Team comfortable with code

### Phase 2: Canary (Week 3)

**Objectives:**
- Enable fallback for small user group
- Validate fallback path works
- Monitor performance

**Configuration:**
```javascript
PD_FALLBACK_ENABLED = true
PD_FALLBACK_TIMEOUT_MS = 5000
```

**User Selection:**
- 5-10 internal users
- Users with cards (some with certs, some without)
- Users willing to provide feedback

**Tasks:**
- [ ] Enable feature flag for canary users
- [ ] Provide extension installation guide
- [ ] Monitor telemetry (success/fail rates, source distribution)
- [ ] Collect user feedback
- [ ] Fix any issues discovered

**Success Criteria:**
- Fallback works for users without certs
- No increase in error rates
- Response times acceptable (< 5s total)
- User feedback positive

**Rollback Trigger:**
- Error rate increase > 10%
- Critical security issue
- User complaints about UX

### Phase 3: Gradual Rollout (Weeks 4-6)

**Objectives:**
- Expand to larger user base
- Validate at scale
- Optimize performance

**Rollout Schedule:**
- **Week 4**: 10% of users
- **Week 5**: 25% of users
- **Week 6**: 50% of users

**Configuration:**
```javascript
PD_FALLBACK_ENABLED = true  // For selected users
PD_FALLBACK_TIMEOUT_MS = 5000
```

**Tasks:**
- [ ] Enable feature flag per user segment
- [ ] Monitor telemetry closely
- [ ] Track fallback usage vs cert usage
- [ ] Optimize timeout if needed
- [ ] Document common issues

**Success Criteria:**
- Error rates stable across segments
- Fallback usage as expected
- No performance degradation
- Support tickets manageable

**Rollback Trigger:**
- Error rate increase > 5% in any segment
- Performance degradation
- Security concerns

### Phase 4: Full Rollout (Week 7+)

**Objectives:**
- Enable for all users
- Monitor long-term stability
- Document lessons learned

**Configuration:**
```javascript
PD_FALLBACK_ENABLED = true  // For all users
PD_FALLBACK_TIMEOUT_MS = 5000  // Or optimized value
```

**Tasks:**
- [ ] Enable feature flag globally
- [ ] Monitor telemetry for 2 weeks
- [ ] Update user documentation
- [ ] Train support team
- [ ] Document rollout learnings

**Success Criteria:**
- Stable error rates
- Fallback used appropriately
- User satisfaction maintained
- No critical issues

## Feature Flag Strategy

### Implementation

Feature flag is controlled via `webapp/config/pd-fallback.config.js`:

```javascript
export const PD_FALLBACK_ENABLED = true;  // or false
```

### Deployment Options

**Option A: Static Config (Current)**
- Change config file, redeploy
- Simple, but requires deployment
- Good for phased rollout

**Option B: Dynamic Config (Future)**
- Load from API endpoint
- Enable/disable without deployment
- Better for rapid rollback

**Recommendation:** Start with Option A, migrate to Option B if needed.

## Telemetry & Monitoring

### Metrics to Track

**Success Metrics:**
- Total identity verification attempts
- Cert path success rate
- Fallback path success rate
- Average response time (cert vs fallback)
- Source distribution (`cert` vs `fallback:chip`)

**Error Metrics:**
- Cert path failures (by error type)
- Fallback path failures (by error type)
- Timeout occurrences
- Extension not installed rate

**Performance Metrics:**
- P50, P95, P99 response times
- Fallback trigger rate
- Card read duration

### Telemetry Implementation

```javascript
// In page-api-auto.js or similar
async function getIdentityAuto() {
  const startTime = Date.now();
  const res = await getIdentityAutoInternal();
  const duration = Date.now() - startTime;
  
  trackEvent('identity_verification', {
    success: res?.ok || false,
    source: res?.data?.source || 'unknown',
    duration: duration,
    error_type: res?.error?.substring(0, 50) || null  // Truncated, no PII
  });
  
  return res;
}
```

**Important:** Never log PD values (surname, givenName, personalCode).

## Rollback Procedures

### Immediate Rollback (< 5 minutes)

**Trigger:**
- Critical security issue
- Data breach suspected
- Widespread failures

**Steps:**
1. Set `PD_FALLBACK_ENABLED = false` in config
2. Deploy config change immediately
3. Verify feature disabled
4. Investigate issue
5. Communicate to users if needed

### Gradual Rollback (< 1 hour)

**Trigger:**
- Error rate increase > 10%
- Performance degradation
- User complaints

**Steps:**
1. Reduce user percentage (e.g., 50% → 25% → 10%)
2. Monitor for improvement
3. If no improvement, disable completely
4. Investigate root cause
5. Fix and retest before re-enabling

### Planned Rollback (Maintenance)

**Trigger:**
- Scheduled maintenance
- Extension update required
- Native helper update required

**Steps:**
1. Notify users in advance
2. Disable feature flag
3. Perform maintenance
4. Re-enable after verification

## Communication Plan

### Internal

- **Week 1**: Announce feature to team
- **Week 2**: Share canary results
- **Week 3-6**: Weekly status updates
- **Week 7**: Rollout complete announcement

### External (if needed)

- **Pre-rollout**: User documentation update
- **During rollout**: Support team training
- **Post-rollout**: User guide updates

## Success Criteria (Overall)

- [ ] Fallback works for users without certificates
- [ ] No regressions in cert path
- [ ] Error rates stable or improved
- [ ] Response times acceptable (< 5s)
- [ ] User satisfaction maintained
- [ ] No security incidents
- [ ] Support tickets manageable

## Post-Rollout

### Week 8+ Activities

- [ ] Analyze telemetry data
- [ ] Document lessons learned
- [ ] Optimize timeout values if needed
- [ ] Plan NFC backend (optional)
- [ ] Consider additional PD fields

### Long-Term Monitoring

- Monitor error rates monthly
- Track fallback usage trends
- Collect user feedback quarterly
- Update documentation as needed

