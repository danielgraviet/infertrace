import grpc
from concurrent import futures
import image_service_pb2
import image_service_pb2_grpc

class ClassifierServicer(image_service_pb2_grpc.ImageClassifierServicer):
    def Predict(self, request, context):
        # 1. Access the raw bytes directly
        img_bytes = request.image_data
        
        # 2. (Simulated) Run your ML model inference
        # result = my_model.predict(img_bytes)
        print(f"Running inference for version: {request.model_version}")
        
        # 3. Return the structured response
        return image_service_pb2.ClassifyResponse(
            label="Golden Retriever",
            confidence=0.98
        )

# Start the server
server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
image_service_pb2_grpc.add_ImageClassifierServicer_to_server(ClassifierServicer(), server)
server.add_insecure_port('[::]:50051')
server.start()