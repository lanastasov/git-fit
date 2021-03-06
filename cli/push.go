package cli

import (
    "github.com/dailymuse/git-fit/transport"
    "github.com/dailymuse/git-fit/config"
    "github.com/dailymuse/git-fit/util"
)

func upload(trans transport.Transport, path string, responseChan chan operationResponse) {
    hash, err := util.FileHash(path)

    if err != nil {
        responseChan <- newOperationResponse(path, transport.NewProgressMessage(0, 0, err), nil)
    } else {
        blob := transport.NewBlob(hash)

        if !util.IsFile(blob.Path()) {
            if err = util.CopyFile(path, blob.Path()); err != nil {
                responseChan <- newOperationResponse(path, transport.NewProgressMessage(0, 0, err), nil)
                return
            }
        }

        exists, err := trans.Exists(blob)

        if err != nil {
            responseChan <- newOperationResponse(path, transport.NewProgressMessage(0, 0, err), nil)
        } else if exists {
            responseChan <- newOperationResponse(path, transport.NewProgressMessage(0, 0, transport.ErrProgressCompleted), blob.Hash)
        } else {
            progress := pipeProgress(path, trans.Upload(blob), responseChan)

            if progress.IsErrored() {
                responseChan <- newOperationResponse(path, progress, nil)
            } else {
                responseChan <- newOperationResponse(path, progress, blob.Hash)
            }
        }
    }
}

func Push(schema *config.Config, trans transport.Transport, args []string) {
    paths := args

    if len(paths) == 0 {
        paths = make([]string, 0)

        for path := range schema.Files {
            paths = append(paths, path)
        }
    }

    responseChan := make(chan operationResponse, len(paths))

    for _, path := range paths {
        go upload(trans, path, responseChan)
    }

    successful := handleResponse(responseChan, len(paths))

    for _, status := range successful {
        if status.Payload != nil {
            schema.Files[status.Path] = status.Payload.(string)
        }
    }
}
