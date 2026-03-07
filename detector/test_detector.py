from detector import Alert, LatencySummary, check


def test_fires_alert_when_p95_doubles():
    # arrange
    baseline = LatencySummary(model_name="gpt-4o", sample_count=100, p95=100.0)
    current  = LatencySummary(model_name="gpt-4o", sample_count=100, p95=210.0)
    # act
    alert = check(current, baseline)
    # assert
    assert alert is not None
    assert alert.model_name == "gpt-4o"
    assert alert.ratio >= 2.0


def test_no_alert_when_below_threshold():
    baseline = LatencySummary(model_name="gpt-4o", sample_count=100, p95=100.0)
    current  = LatencySummary(model_name="gpt-4o", sample_count=100, p95=150.0)
    alert = check(current, baseline)
    assert alert is None


def test_alert_fires_at_exact_threshold():
    baseline = LatencySummary(model_name="gpt-4o", sample_count=100, p95=100.0)
    current  = LatencySummary(model_name="gpt-4o", sample_count=100, p95=200.0)
    alert = check(current, baseline)
    assert alert is not None


def test_no_alert_when_baseline_p95_is_zero():
    baseline = LatencySummary(model_name="gpt-4o", sample_count=0, p95=0.0)
    current  = LatencySummary(model_name="gpt-4o", sample_count=100, p95=500.0)
    alert = check(current, baseline)
    assert alert is None


def test_no_alert_when_baseline_p95_is_negative():
    baseline = LatencySummary(model_name="gpt-4o", sample_count=0, p95=-1.0)
    current  = LatencySummary(model_name="gpt-4o", sample_count=100, p95=500.0)
    alert = check(current, baseline)
    assert alert is None
