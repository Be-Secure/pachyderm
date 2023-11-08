import os
from pathlib import Path
from distutils.util import strtobool

HTTP_UNIX_SOCKET_SCHEMA = "http_unix"
HTTP_SCHEMA = "http"
DEFAULT_SCHEMA = os.environ.get("DEFAULT_SCHEMA", HTTP_UNIX_SOCKET_SCHEMA)
SIDECAR_MODE = strtobool(os.environ.get("SIDECAR_MODE", "False").lower())
# TODO(msteffen) Maybe depend on Pachyderm SDK for this
PACH_CONFIG = os.environ.get("PACH_CONFIG", Path.home() / ".pachyderm/config.json")
DET_RESOURCES_TYPE = os.environ.get("DET_RESOURCES_TYPE")  # Assume MLDE if set
SLURM_JOB = "slurm-job"  # For launcher this is set in DispatcherRM

PACHYDERM_EXT_DEBUG = strtobool(os.environ.get("PACHYDERM_EXT_DEBUG", "False").lower())
if PACHYDERM_EXT_DEBUG:
    from jupyterlab_pachyderm.log import get_logger

    logger = get_logger()
    logger.setLevel("DEBUG")
    logger.debug("DEBUG mode activated for pachyderm extension")
