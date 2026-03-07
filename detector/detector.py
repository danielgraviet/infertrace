# this is a module that just takes in requests, and is able to show us if there has been a spike. 
# return enough info so that the user knows what to do, and can make an informed decision
# Questions
# - should the be a separate return object that we pass to the user? this data model would contain important info like the model, time, etc. 
# - how do we know when to use pydantic in a project? there is a new data validation library I want to try out that is more light weight called frfr
# should our alert class be a separate module, or defined in this detector.py file
from dataclasses import dataclass

@dataclass
class LatencySummary:
    model_name: str
    sample_count: int
    p95: float

@dataclass
class Alert:
    model_name: str
    current_p95: float
    baseline_p95: float
    ratio: float
    

def check(current: LatencySummary, baseline: LatencySummary, threshold: float = 2.0) -> Alert | None:
    if baseline.p95 <= 0: # prevent div by 0 err.
        return None
    
    ratio = current.p95 / baseline.p95
    if ratio >= threshold:
        return Alert(
            model_name=current.model_name,
            current_p95=current.p95,
            baseline_p95=baseline.p95,
            ratio=ratio
        )
    return None

