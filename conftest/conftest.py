import pytest
import signal

total_points = 0
earned_points = 0


def pytest_configure(config):
    config.addinivalue_line(
        "markers", "grade(points, feedback): assign points and feedback to a test"
    )


def pytest_runtest_makereport(item, call):
    global total_points, earned_points

    if call.when != "call":
        return

    marker = item.get_closest_marker("grade")
    if marker:
        points = marker.kwargs.get("points", 0)
        feedback = marker.kwargs.get("feedback", "")

        total_points += points

        if call.excinfo is None:
            earned_points += points
            print(f"\n✔ {item.name} ({points}/{points}) - {feedback}")
        else:
            print(f"\n✘ {item.name} (0/{points}) - {feedback}")


@pytest.fixture(autouse=True)
def timeout():
    """Handle possible infinite loops"""

    def handler(signum, frame):
        raise TimeoutError("Test ran for more than 5 seconds - possible infinite loop")

    signal.signal(signal.SIGALRM, handler)
    signal.alarm(5)
    yield
    signal.alarm(0)
