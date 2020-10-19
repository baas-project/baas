import asyncio
from concurrent import futures
import grpc
from servicea import main_pb2_grpc
from servicea import main_impl


def main():
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    main_pb2_grpc.add_GreeterServicer_to_server(main_impl.Greeter(), server)
    server.add_insecure_port('[::]:50051')
    server.start()

    server.wait_for_termination()


if __name__ == '__main__':
    main()
    # loop = asyncio.get_event_loop()
    # loop.create_task(main())
    # loop.run_forever()
