import asyncio


async def main():
    print("Hello from python!")


loop = asyncio.get_event_loop()
loop.create_task(main())
loop.run_forever()
