import datetime
import json
import time
from typing import Dict

from pytz import utc
from requests import Session

_EXPIRY = datetime.timedelta(hours=24)


def get_job() -> Dict:
    created_at = datetime.datetime.utcnow().replace(tzinfo=utc)
    expires_at = created_at + _EXPIRY

    return {
        "name": "job1",
        "schedule": "@every 1s",
        "timezone": "Australia/Perth",
        "owner": "Manual Test",
        "owner_email": "no-reply@localhost",
        "disabled": False,
        "tags": {},  # use this to "route" jobs to certain dkron nodes
        "metadata": {
            "some": "metadata",
        },
        "concurrency": "forbid",
        "executor": "nats",
        "executor_config": {
            "url": "nats://nats:4222",
            "message": json.dumps(
                {
                    "created_at": created_at.isoformat(),
                    "expires_at": expires_at.isoformat(),
                    "url": "https://random-data-api.com/api/cannabis/random_cannabis?size=1",
                }
            ),
            "subject": "scrape_jobs",
            "userName": "sandbox",
            "password": "s@nb0x123!@#",
        },
    }


# to check: nats sub -s host.docker.internal:4222 scrape_jobs
def main():
    with Session() as s:
        job = get_job()

        r = s.post(
            url="http://host.docker.internal:8181/v1/jobs",
            json=job,
            timeout=5,
        )
        print(r)
        print(r.text)

        print("")
        print("Ctrl + C to quit...")
        while 1:
            try:
                time.sleep(1)
            except KeyboardInterrupt:
                break

        r = s.delete(
            f"http://host.docker.internal:8181/v1/jobs/{job['name']}",
            timeout=5,
        )
        print(r)
        print(r.text)


if __name__ == "__main__":
    main()
