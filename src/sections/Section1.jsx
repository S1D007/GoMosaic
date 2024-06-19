import { useEffect, useState } from "react";
import { open } from "@tauri-apps/api/dialog";
import axios from "axios";
import { toast } from "react-toastify";

export default function Section1() {
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(false);
  const [downloading, setDownloading] = useState(false);
  const [formData, setFormData] = useState({
    rows: 0,
    cols: 0,
    width: 0,
    height: 0,
    output: "",
    image: null,
  });
  const [formData2, setFormData2] = useState({
    gridCellFolder: "",
    inputFolder: "",
    outputFolder: "",
    opacity: 0.6,
  });

  const handleSelectFolder = async (folder) => {
    try {
      const selectedPath = await open({
        directory: true,
        multiple: false,
      });
      if (selectedPath) {
        if (folder == "output")
          setFormData({ ...formData, output: selectedPath });
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
        rows: 0,
        cols: 0,
        width: 0,
        height: 0,
        output: "",
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
      link.download = "downloaded_image.png";
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      toast.success("Backdrop Image Downloaded successfully!");
    } catch (error) {
      console.error("Error:", error.response);
      toast.error("Error While Downloading Backdrop Image. Please try again.");
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
        opacity: 0.6,
      });
    } catch (error) {
      console.error("Error:", error);
      toast.error("Error While Starting the Server. Please try again.");
    } finally {
      setLoading(false);
    }
  };

  useEffect(()=>{console.log(formData)},[formData])

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
  

  return (
    <div
      style={{ display: "flex", justifyContent: "center", padding: "4rem 0 0 0" }}
    >
      {page==2&&<div onClick={()=>setPage(1)} style={{position:"absolute", top:10, left:10, display:"flex", alignItems:"center", justifyContent:"center", cursor:"pointer", fontSize:"1rem"}}>&#x2190; Go Back</div>}
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
              Enter Number of Rows
            </p>
            <input
              type="number"
              onChange={(e) =>
                setFormData({ ...formData, rows: parseInt(e.target.value) })
              }
            />
          </div>
          <div>
            <p style={{ color: "rgb(189, 189, 189)", margin: 3 }}>
              Enter Number of Columns
            </p>
            <input
              type="number"
              onChange={(e) =>
                setFormData({ ...formData, cols: parseInt(e.target.value) })
              }
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
              {formData.image?.name}
            </p>
          </div>
          <div>
            <button
              style={{ backgroundColor: "#0F52BA", width: "236px", color:"whitesmoke" }}
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
              {formData.output}
            </p>
          </div>
          <div>
            <button
              style={{ backgroundColor: "#0F52BA", width: "236px", color:"whitesmoke" }}
              onClick={cutImageIntoGrid}
              disabled={loading}
            >
              {loading ? "Please Wait ..." : "Cut Images Into Grid"}
            </button>
          </div>
          <div>
            <button
              style={{ backgroundColor: "#0F52BA", width: "236px", color:"whitesmoke" }}
              onClick={()=>setPage(2)}
            >
              Start Mosaic Server
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
              onChange={(e) =>
                setFormData2({ ...formData2, opacity: e.target.value })
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
              {formData2.gridCellFolder}
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
              {formData2.inputFolder}
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
              {formData2.outputFolder}
            </p>
          </div>
          <button
            style={{ width: "236px", backgroundColor: "#0F52BA", color:"whitesmoke" }}
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
            style={{ width: "236px", backgroundColor: "#0F52BA", color:"white" }}
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
