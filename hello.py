import os
import subprocess
import sys
from datetime import datetime, timezone

if "--verbose" in sys.argv[1:]:
    print(datetime.now(timezone.utc).isoformat().replace("+00:00", "Z"))
print("hello world")

if os.environ.get("HELLO_SELF_RUN") != "1":
    environment = os.environ.copy()
    environment["HELLO_SELF_RUN"] = "1"
    subprocess.run([sys.executable, __file__, *sys.argv[1:]], env=environment, check=True)
