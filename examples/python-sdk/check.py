"""Run from the repository root against a reviewed configuration and matching CLI."""
import argparse
import asyncio

from agenthealth import AgentHealthError, Client


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('configuration')
    parser.add_argument('--binary', default='agenthealth')
    parser.add_argument('--async', dest='asynchronous', action='store_true')
    args = parser.parse_args()
    client = Client(args.binary)
    try:
        report = (asyncio.run(client.async_check(args.configuration)) if args.asynchronous
                  else client.check(args.configuration))
    except AgentHealthError as error:
        print(str(error))
        return 6
    print(report.status.value)
    for root in report.results:
        print(root.target.name, root.status.value)
        for dependency in root.dependencies:
            print(dependency.relationship, dependency.target.name, dependency.status.value)
    return report.exit_code


if __name__ == '__main__':
    raise SystemExit(main())
