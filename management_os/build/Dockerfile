FROM golang as build

WORKDIR /build


# Force modules
ENV GO111MODULE=on

# Cache dependencies
COPY go.* ./
RUN go mod download

# Build project
COPY . .
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" ./management_os/programs

# ---------------------

FROM archlinux
RUN pacman -Syu dhcpcd --noconfirm
RUN systemctl enable getty@tty1.service
RUN systemctl enable dhcpcd
RUN ln -sr /usr/lib/systemd/systemd /init

RUN echo -e "toor\ntoor" | passwd
RUN ln -s /usr/share/zoneinfo/Europe/Amsterdam /etc/localtime

COPY --from=build /build/programs programs
