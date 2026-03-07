import time
import requests
from detector import Alert, LatencySummary, check

BASE_URL = "http://localhost:8080"
POLL_INTERVAL_SECONDS = 30
MODELS_TO_WATCH = ["gpt-4o", "llama-3"]

CURRENT_WINDOW = "5m"
BASELINE_WINDOW = "30m"


def fetch_summary(model_name: str, window: str) -> LatencySummary | None:
    url = f"{BASE_URL}/models/{model_name}/latency?window={window}"
    response = requests.get(url, timeout=5)
    if response.status_code != 200:
        print(f"[poller] failed to fetch {model_name} ({window}): {response.status_code}")
        return None

    data = response.json()
    return LatencySummary(
        model_name=model_name,
        sample_count=data["sample_count"],
        p95=data["p95"],
    )


def on_alert(alert: Alert) -> None:
    print(
        f"[ALERT] {alert.model_name} | "
        f"current_p95={alert.current_p95:.0f}ns | "
        f"baseline_p95={alert.baseline_p95:.0f}ns | "
        f"ratio={alert.ratio:.2f}x"
    )


def poll_once() -> None:
    for model_name in MODELS_TO_WATCH:
        current = fetch_summary(model_name, CURRENT_WINDOW)
        baseline = fetch_summary(model_name, BASELINE_WINDOW)

        if current is None or baseline is None:
            continue

        alert = check(current, baseline)
        if alert:
            on_alert(alert)


def run() -> None:
    print(f"[poller] starting — watching {MODELS_TO_WATCH} every {POLL_INTERVAL_SECONDS}s")
    while True:
        poll_once()
        time.sleep(POLL_INTERVAL_SECONDS)


if __name__ == "__main__":
    run()
