import infertrace
import time

infertrace.init("localhost:4317", "test-service")

@infertrace.trace(model="gpt-4")
def fake_inference():
    time.sleep(0.1)

fake_inference()
print("span sent")
