import inspect

from typing import List
from types import FunctionType, BuiltinFunctionType

from flask import request, has_request_context

from opentelemetry import trace
from opentelemetry.exporter.otlp.proto.grpc.trace_exporter import OTLPSpanExporter
from opentelemetry.sdk.resources import SERVICE_NAME, Resource
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.propagators.textmap import DefaultGetter
from opentelemetry.trace.propagation.tracecontext import TraceContextTextMapPropagator


class Trace:
    def __init__(
        self,
        disable_tracing: bool = False,
    ):
        self.disable_tracing = disable_tracing
        self.tracer = trace.get_tracer(__name__)

    def _already_decorated_methods(self, target_class, decorator_name) -> List[str]:
        methods = []
        sourcelines = inspect.getsourcelines(target_class)[0]
        for i, line in enumerate(sourcelines):
            line = line.strip()
            line = line.split("(")
            if len(line) > 0:
                if line[0].strip() == f"@{decorator_name}":
                    nextLine = sourcelines[i + 1]
                    nextLine = nextLine.split("def")
                    if len(nextLine) == 2:
                        name = nextLine[1].split("(")[0].strip()
                        methods.append(name)

        return methods

    def _tracing_trace_decorator(self, method):
        return method.__name__ == "wrapped_f"

    def __call__(self, func_or_class):
        if self._tracing_trace_decorator(
            func_or_class
        ):  # we want to avoid tracing trace decorator itself
            return func_or_class

        # in case class was decorated we want to wrap all its methods into Trace() decorator
        if isinstance(func_or_class, type):
            # at this point we already know that this is a class
            cls = func_or_class

            # as we want to be able to overwrite Trace decorator behaviour per function,
            # we have to avoid decorate methods twice
            already_decorated = self._already_decorated_methods(
                cls, self.__class__.__name__
            )

            # here we retrieve all methods of the given class and decorate all eligible
            for name, method in vars(cls).items():
                if name not in already_decorated:
                    trace_decorator = Trace(
                        disable_tracing=self.disable_tracing,
                    )

                    if isinstance(method, (FunctionType, BuiltinFunctionType)):
                        setattr(cls, name, trace_decorator(method))

                    # a special case, when we want to decorate class/static method
                    #  we have to follow the next order of decorators @classmethod(@Trace()(func))
                    elif isinstance(method, (classmethod, staticmethod)):
                        inner_func = method.__func__
                        method_type = type(method)
                        setattr(cls, name, method_type(trace_decorator(inner_func)))

            return cls

        def wrapped_f(*args, **kwargs):
            # at this point we know that this is a function
            fn = func_or_class

            if has_request_context():

                def ctx():
                    # as we want to propagate request from the upstream only on the initial request,
                    # we have to check flag if it was already porpagated
                    if request.environ.get("DO_NOT_PROPAGATE_CONTEXT"):
                        return None

                    # we extract headers values, because if seggestions microservice
                    # was called from another instrumented service,
                    # we want to have a fancy dependency graph in traces explorer
                    return TraceContextTextMapPropagator().extract(
                        getter=DefaultGetter(), carrier=request.headers
                    )

                if self.disable_tracing:
                    # we have to propagate tracing disabling to all underlying components
                    request.environ["DISABLE_TRACING"] = self.disable_tracing

                    return fn(*args, **kwargs)

                # if there is a DISABLE_TRACING env variable in the request,
                #   it means that upperlevel component propagated tracing disablement on uderlying components
                if request.environ.get("DISABLE_TRACING"):
                    return fn(*args, **kwargs)

                # here we sure that tracing is enabled and we are ready to instrument the code
                span_name = f"{request.path} ({fn.__qualname__})"

                with self.tracer.start_as_current_span(span_name, context=ctx()) as _:
                    # when the first request was made we have to flag request with
                    # do not propagate initial context anymore,
                    # because we want to save dependencies inside the service,
                    # and do not just reuse the initial context every time
                    request.environ["DO_NOT_PROPAGATE_CONTEXT"] = True

                    return fn(*args, **kwargs)

            return fn(*args, **kwargs)

        return wrapped_f


class OTLPProvider:
    def __init__(
        self,
        service_name: str,
        exporter_endpoint: str,
    ) -> None:
        self._provider = TracerProvider(
            resource=Resource(attributes={SERVICE_NAME: service_name}),
            active_span_processor=BatchSpanProcessor(
                OTLPSpanExporter(
                    endpoint=exporter_endpoint,
                    insecure=True,
                )
            ),
        )

    @property
    def provider(self) -> TracerProvider:
        return self._provider
