import { useEffect, useState } from "react";
import { open } from "@tauri-apps/api/dialog";
import axios from "axios";
import { toast } from "react-toastify";
import Box from "@mui/material/Box";
import Typography from "@mui/material/Typography";
import Modal from "@mui/material/Modal";

const style = {
  position: "absolute",
  top: "50%",
  left: "50%",
  transform: "translate(-50%, -50%)",
  width: 500,
  bgcolor: "background.paper",
  border: "2px solid #000",
  boxShadow: 24,
  p: 4,
};

export default function Section1() {
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(false);
  const [downloading, setDownloading] = useState(false);
  const [formData, setFormData] = useState({
    pax: 0,
    rows: 0,
    cols: 0,
    width: 0,
    height: 0,
    gridCellHeight: 0,
    gridCellWidth: 0,
    output: "",
    backdrop: "",
    image: null,
  });
  const [formData2, setFormData2] = useState({
    gridCellFolder: "",
    inputFolder: "",
    outputFolder: "",
    opacity: "",
  });
  const [result, setResult] = useState(null);
  const [paxTest, setPaxTest] = useState("");
  const [vary, setVary] = useState(10);
  const [openModal, setOpen] = useState(false);

  const handleSelectFolder = async (folder) => {
    try {
      const selectedPath = await open({
        directory: true,
        multiple: false,
      });
      if (selectedPath) {
        if (folder == "output")
          setFormData({ ...formData, output: selectedPath });
        else if (folder == "backdrop")
          setFormData({ ...formData, backdrop: selectedPath });
        else if (folder == "gridCellFolder")
          setFormData2({ ...formData2, gridCellFolder: selectedPath });
        else if (folder == "inputFolder")
          setFormData2({ ...formData2, inputFolder: selectedPath });
        else if (folder == "outputFolder")
          setFormData2({ ...formData2, outputFolder: selectedPath });
      }
    } catch (error) {
      console.error("Error selecting folder:", error);
    }
  };

  const cutImageIntoGrid = async () => {
    setLoading(true);
    try {
      const formDataToSend = new FormData();
      formDataToSend.append("rows", formData.rows);
      formDataToSend.append("cols", formData.cols);
      formDataToSend.append("output", formData.output);
      formDataToSend.append("image", formData.image);
      const response = await axios.post(
        "http://127.0.0.1:8000/mosaic",
        formDataToSend
      );
      console.log("Response:", response.data);
      toast.success(
        "Images Generated Successfully ! Check Your Selected Grid Cell Folder"
      );
      setFormData({
        pax: 0,
        rows: 0,
        cols: 0,
        width: 0,
        height: 0,
        gridCellHeight: 0,
        gridCellWidth: 0,
        output: "",
        backdrop: "",
        image: null,
      });
    } catch (error) {
      console.error("Error:", error);
      toast.error("Error While Cutting the Image. Please try again.");
    } finally {
      setLoading(false);
    }
  };

  const downloadBackdrop = async () => {
    setDownloading(true);
    try {
      const response = await axios.post(
        "http://localhost:8000/backdrop",
        {
          rows: formData.rows,
          cols: formData.cols,
          width: formData.width,
          height: formData.height,
        },
        {
          responseType: "arraybuffer",
        }
      );
      const blob = new Blob([response.data], { type: "image/png" });
      const link = document.createElement("a");
      const url = URL.createObjectURL(blob);
      link.href = url;
      link.download = "Backdrop.png";
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      toast.success("Backdrop Image Downloaded successfully!");
    } catch (error) {
      console.error("Error:", error);
      // toast.error(request.response.data);
    } finally {
      setDownloading(false);
    }
  };

  const startServer = async () => {
    setLoading(true);
    try {
      const formDataToSend = new FormData();
      formDataToSend.append("gridCellFolder", formData2.gridCellFolder);
      formDataToSend.append("inputFolder", formData2.inputFolder);
      formDataToSend.append("outputFolder", formData2.outputFolder);
      formDataToSend.append("opacity", formData2.opacity);
      const response = await axios.post(
        "http://127.0.0.1:8000/start-overlay",
        formDataToSend
      );
      console.log("Response:", response.data);
      toast.success("Server Is Ready!");
      setPage(3);
      setFormData2({
        gridCellFolder: "",
        inputFolder: "",
        outputFolder: "",
        opacity: "",
      });
    } catch (error) {
      console.error("Error:", error);
      toast.error("Error While Starting the Server. Please try again.");
    } finally {
      setLoading(false);
    }
  };

  const handleImageUpload = (e) => {
    const file = e.target.files[0];
    if (file) {
      const reader = new FileReader();
      reader.onload = (event) => {
        const img = new Image();
        img.onload = () => {
          setFormData((prevFormData) => ({
            ...prevFormData,
            image: file,
            width: img.width,
            height: img.height,
          }));
        };
        img.src = event.target.result;
      };
      reader.readAsDataURL(file);
    }
  };

  const cutter = () => {
    let results = [];
    let vis = new Map();
    let i = 1;
    while (i <= formData.width && i <= formData.height) {
      const rows = Math.floor(formData.height / i);
      const columns = Math.floor(formData.width / i);
      const tmpPax = rows * columns;

      if (tmpPax >= paxTest - vary && tmpPax <= paxTest + vary) {
        const key = `${rows},${columns}`;
        if (!vis.has(key)) {
          vis.set(key, 1);
          results.push({
            gridCellWidth: formData.width / columns,
            gridCellHeight: formData.height / rows,
            rows: rows,
            columns: columns,
            pax: rows * columns,
          });
        }
      }
      i++;
    }
    const filtered = results.filter(
      (result) => result.pax >= paxTest - 10 && result.pax <= paxTest + 10
    );
    setResult(filtered);
  };

  return (
    <div
      style={{
        display: "flex",
        justifyContent: "center",
        padding: "4rem 0 1rem 0",
      }}
    >
      <button
        style={{
          position: "absolute",
          top: 70,
          right: 20,
          width: "236px",
          backgroundColor: "#cd3b3b",
          color: "whitesmoke",
        }}
        onClick={() => {
          setFormData2({
            gridCellFolder: "",
            inputFolder: "",
            outputFolder: "",
            opacity: "",
          });
          setFormData({
            pax: 0,
            rows: 0,
            cols: 0,
            width: 0,
            height: 0,
            gridCellHeight: 0,
            gridCellWidth: 0,
            output: "",
            backdrop: "",
            image: null,
          });
          setResult(null);
          setPaxTest("");
        }}
      >
        Reset
      </button>
      <button
        style={{
          position: "absolute",
          top: 120,
          borderBottomLeftRadius:0,
          borderBottomRightRadius:0,
          right: 20,
          backgroundColor: page==1?'#0F52BA':'gray',
          width: "236px",
          color: "whitesmoke",
        }}
        onClick={() => {
         setPage(1);
        }}
      >
        Cut Image To Grid
      </button>
      <button
        style={{
          position: "absolute",
          top: 161,
          borderTopLeftRadius:0,
          borderTopRightRadius:0,
          right: 20,
          backgroundColor:  page==2?'#0F52BA':'gray',
          width: "236px",
          color: "whitesmoke",
        }}
        onClick={() => {
        setPage(2);
        }}
      >
        Start Server
      </button>
      {page === 1 && (
        <div
          style={{
            display: "flex",
            flexDirection: "column",
            rowGap: "20px",
            justifyContent: "center",
          }}
        >
          <div>
            <p style={{ color: "rgb(189, 189, 189)", margin: 3 }}>
              Enter Number of Users
            </p>
            <input
              type="number"
              value={paxTest}
              onChange={(e) => setPaxTest(parseInt(e.target.value))}
            />
          </div>
          <div>
            <input
              type="file"
              id="fileInput"
              onChange={handleImageUpload}
              style={{ display: "none" }}
            />
            <button
              type="button"
              style={{ width: "236px" }}
              onClick={() => document.getElementById("fileInput").click()}
            >
              Choose Mosaic Image
            </button>
            <p
              style={{
                color: "rgb(189, 189, 189)",
                padding: 0,
                fontSize: "14px",
                margin: 3,
              }}
            >
              {formData.image?.name.length > 30
                ? formData.image?.name.slice(0, 30) + "..."
                : formData.image?.name}
            </p>
          </div>
          <div style={{ color: "rgb(189, 189, 189)", fontSize: "0.9rem" }}>
            <div>Image Width : {formData.width} px</div>
            <div>Image Height : {formData.height} px</div>
          </div>
          <div>
            <button
              style={{
                backgroundColor: "#0F52BA",
                width: "236px",
                color: "whitesmoke",
              }}
              onClick={() => {
                cutter();
                setOpen(true);
              }}
            >
              Get Possibilities
            </button>
            <Modal
              open={openModal}
              onClose={() => setOpen(false)}
              aria-labelledby="modal-modal-title"
              aria-describedby="modal-modal-description"
            >
              <Box sx={style}>
                <Typography
                  id="modal-modal-description"
                  sx={{
                    mt: 2,
                    display: "flex",
                    flexDirection: "column",
                    rowGap: 3,
                  }}
                >
                  {result?.map((r, index) => (
                    <button
                      key={index}
                      style={{
                        backgroundColor: "#0F52BA",
                        color: "whitesmoke",
                      }}
                      onClick={() => {
                        setFormData({
                          ...formData,
                          rows: r.rows,
                          cols: r.columns,
                          pax: r.pax,
                          gridCellHeight: r.gridCellHeight,
                          gridCellWidth: r.gridCellWidth,
                        });
                        setResult(null);
                        setOpen(false);
                      }}
                    >
                      Rows: <b>{r.rows}</b>, Cols: <b>{r.columns}</b>, Pax:{" "}
                      <b>{r.pax}</b>, Cell Size:{" "}
                      <b>{r.gridCellWidth.toFixed(2)}</b> x{" "}
                      <b>{r.gridCellHeight.toFixed(2)}</b> px
                    </button>
                  ))}
                </Typography>
              </Box>
            </Modal>
          </div>
          <div style={{ color: "rgb(189, 189, 189)", fontSize: "0.9rem" }}>
            <div>No Of Rows : {formData.rows}</div>
            <div>No Of Columns : {formData.cols}</div>
            <div>No Of Pax : {formData.pax}</div>
            <div>
              Cell Width :{" "}
              {`${formData.gridCellWidth.toFixed(
                2
              )} x ${formData.gridCellHeight.toFixed(2)} px`}
            </div>
          </div>
          <div>
            <button
              style={{ width: "236px" }}
              onClick={() => handleSelectFolder("backdrop")}
            >
              Select Backdrop Folder
            </button>
            <p
              style={{
                color: "rgb(189, 189, 189)",
                margin: 3,
                padding: 0,
                fontSize: "14px",
              }}
            >
              {formData.backdrop.length > 30
                ? formData.backdrop.slice(0, 30) + "..."
                : formData.backdrop}
            </p>
          </div>
          <div>
            <button
              style={{
                backgroundColor: "#0F52BA",
                width: "236px",
                color: "whitesmoke",
              }}
              onClick={downloadBackdrop}
              disabled={downloading}
            >
              {downloading ? "Please Wait ..." : "Download Backdrop Image"}
            </button>
          </div>
          <div>
            <button
              style={{ width: "236px" }}
              onClick={() => handleSelectFolder("output")}
            >
              Select Grid Cell Folder
            </button>
            <p
              style={{
                color: "rgb(189, 189, 189)",
                margin: 3,
                padding: 0,
                fontSize: "14px",
              }}
            >
              {formData.output.length > 30
                ? formData.output.slice(0, 30) + "..."
                : formData.output}
            </p>
          </div>
          <div>
            <button
              style={{
                backgroundColor: "#0F52BA",
                width: "236px",
                color: "whitesmoke",
              }}
              onClick={cutImageIntoGrid}
              disabled={loading}
            >
              {loading ? "Please Wait ..." : "Cut Images Into Grid"}
            </button>
          </div>
        </div>
      )}
      {page === 2 && (
        <div
          style={{
            display: "flex",
            flexDirection: "column",
            rowGap: "20px",
            justifyContent: "center",
          }}
        >
          <div>
            <p style={{ color: "rgb(189, 189, 189)", margin: 3 }}>
              Enter Opacity
            </p>
            <input
              type="number"
              value={formData2.opacity}
              onChange={(e) =>
                setFormData2({
                  ...formData2,
                  opacity: parseInt(e.target.value),
                })
              }
            />
          </div>
          <div>
            <button
              style={{ width: "236px" }}
              onClick={() => handleSelectFolder("gridCellFolder")}
            >
              Select Grid Cell Folder
            </button>
            <p
              style={{
                color: "rgb(189, 189, 189)",
                margin: 3,
                padding: 0,
                fontSize: "14px",
              }}
            >
              {" "}
              {formData2.gridCellFolder.length > 30
                ? formData2.gridCellFolder.slice(0, 30) + "..."
                : formData2.gridCellFolder}
            </p>
          </div>
          <div>
            <button
              style={{ width: "236px" }}
              onClick={() => handleSelectFolder("inputFolder")}
            >
              Select Input Folder
            </button>
            <p
              style={{
                color: "rgb(189, 189, 189)",
                margin: 3,
                padding: 0,
                fontSize: "14px",
              }}
            >
              {" "}
              {formData2.inputFolder.length > 30
                ? formData2.inputFolder.slice(0, 30) + "..."
                : formData2.inputFolder}
            </p>
          </div>
          <div>
            <button
              style={{ width: "236px" }}
              onClick={() => handleSelectFolder("outputFolder")}
            >
              Select Output Folder
            </button>
            <p
              style={{
                color: "rgb(189, 189, 189)",
                margin: 3,
                padding: 0,
                fontSize: "14px",
              }}
            >
              {" "}
              {formData2.outputFolder.length > 30
                ? formData2.outputFolder.slice(0, 30) + "..."
                : formData2.outputFolder}
            </p>
          </div>
          <button
            style={{
              width: "236px",
              backgroundColor: "#0F52BA",
              color: "whitesmoke",
            }}
            onClick={startServer}
            disabled={loading}
          >
            {loading ? "Please Wait ..." : "Start Server"}
          </button>
        </div>
      )}
      {page === 3 && (
        <div
          style={{
            display: "flex",
            flexDirection: "column",
            color: "rgb(189, 189, 189)",
            justifyContent: "center",
            alignItems: "center",
          }}
        >
          <p
            style={{
              display: "flex",
              justifyContent: "center",
              alignSelf: "center",
              alignItems: "center",
              fontSize: "1.2rem",
              textAlign: "center",
              lineHeight: "40px",
            }}
          >
            Server Is Ready !! <br /> Start Adding Pictures in Input Folder
          </p>
          <button
            style={{
              width: "236px",
              backgroundColor: "#0F52BA",
              color: "white",
            }}
            onClick={() => window.location.reload()}
            disabled={loading}
          >
            Regenerate
          </button>
        </div>
      )}
    </div>
  );
}
