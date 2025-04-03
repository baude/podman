package server

import (
	"github.com/containers/podman/v5/pkg/api/handlers/libpod"
	"github.com/gorilla/mux"
	"net/http"
)

func (s *APIServer) registerArtifactHandlers(r *mux.Router) error {
	// swagger:operation GET /libpod/containers/{name}/archive libpod ContainerArchiveLibpod
	// ---
	//  summary: Copy files from a container
	//  description: Copy a tar archive of files from a container
	//  tags:
	//   - containers (compat)
	//  produces:
	//  - application/json
	//  parameters:
	//   - in: path
	//     name: name
	//     type: string
	//     description: container name or id
	//     required: true
	//   - in: query
	//     name: path
	//     type: string
	//     description: Path to a directory in the container to extract
	//     required: true
	//   - in: query
	//     name: rename
	//     type: string
	//     description: JSON encoded map[string]string to translate paths
	//  responses:
	//    200:
	//      description: no error
	//      schema:
	//       type: string
	//       format: binary
	//    400:
	//      $ref: "#/responses/badParamError"
	//    404:
	//      $ref: "#/responses/containerNotFound"
	//    500:
	//      $ref: "#/responses/internalError"
	r.HandleFunc(VersionedPath("/libpod/artifacts/{name}/json"), s.APIHandler(libpod.InspectArtifact)).Methods(http.MethodGet)
	return nil
}
